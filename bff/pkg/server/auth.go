package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type contextKey string

const userContextKey contextKey = "user"

type cachedUser struct {
	info      *UserInfo
	expiresAt time.Time
}

var (
	authCache  sync.Map
	authClient *kubernetes.Clientset
	authOnce   sync.Once
	devMode    = os.Getenv("DEV_MODE") == "true"
)

func initAuthClient() {
	authOnce.Do(func() {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return
		}
		authClient, _ = kubernetes.NewForConfig(cfg)
	})
}

func GetUser(r *http.Request) *UserInfo {
	if user, ok := r.Context().Value(userContextKey).(*UserInfo); ok {
		return user
	}
	return &UserInfo{Username: "", IsAdmin: false}
}

func AuthMiddleware(next http.Handler) http.Handler {
	initAuthClient()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/health") {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			if authClient == nil && devMode {
				ctx := context.WithValue(r.Context(), userContextKey, &UserInfo{Username: "dev", Groups: []string{"system:authenticated"}, IsAdmin: true})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		hash := sha256.Sum256([]byte(token))
		cacheKey := hex.EncodeToString(hash[:])
		if cached, ok := authCache.Load(cacheKey); ok {
			cu := cached.(*cachedUser)
			if time.Now().Before(cu.expiresAt) {
				ctx := context.WithValue(r.Context(), userContextKey, cu.info)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			authCache.Delete(cacheKey)
		}

		if authClient == nil {
			if devMode {
				ctx := context.WithValue(r.Context(), userContextKey, &UserInfo{Username: "dev", IsAdmin: true})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			writeError(w, http.StatusServiceUnavailable, "authentication service unavailable")
			return
		}

		tr, err := authClient.AuthenticationV1().TokenReviews().Create(r.Context(), &authenticationv1.TokenReview{
			Spec: authenticationv1.TokenReviewSpec{Token: token},
		}, metav1.CreateOptions{})
		if err != nil || !tr.Status.Authenticated {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		username := tr.Status.User.Username
		groups := tr.Status.User.Groups
		sar, err := authClient.AuthorizationV1().SubjectAccessReviews().Create(r.Context(), &authorizationv1.SubjectAccessReview{
			Spec: authorizationv1.SubjectAccessReviewSpec{
				User:   username,
				Groups: groups,
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Group:    "finops.maas.io",
					Resource: "plugins",
					Verb:     "admin",
				},
			},
		}, metav1.CreateOptions{})
		isAdmin := err == nil && sar.Status.Allowed

		user := &UserInfo{Username: username, Groups: groups, IsAdmin: isAdmin}
		authCache.Store(cacheKey, &cachedUser{info: user, expiresAt: time.Now().Add(60 * time.Second)})
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
