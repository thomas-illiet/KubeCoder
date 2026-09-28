# Kube Coder

Le plan directeur du projet se trouve dans [`Project/README.md`](./Project/README.md).

## Prototype du panneau d’administration

Le prototype est une application frontend en Vue 3, TypeScript et Vuetify 3. Les données métier affichées restent fictives, mais l’accès au site utilise une authentification OpenID Connect réelle contre l’instance Keycloak locale décrite dans [`deploy/helm/kubecoder-keycloak`](./deploy/helm/kubecoder-keycloak/README.md).

```sh
helm upgrade --install kubecoder-auth ./deploy/helm/kubecoder-keycloak \
  --namespace kubecoder-auth --create-namespace --wait --timeout 10m

cd frontend
cp .env.example .env.local
npm install
npm run dev
```

L'interface est alors disponible sur `http://localhost:5173/admin`.

Le compte local est `admin` / `admin`. Le frontend utilise une bibliothèque OIDC générique, le flux Authorization Code avec PKCE S256 et `sessionStorage`; il ne dépend pas de `keycloak-js`. Toutes les routes organisation et administration exigent une session, sans appliquer de rôle ni d’autorisation métier.

Pour vérifier le build de production :

```powershell
cd frontend
npm run build
```

Les identifiants présents dans le chart et l’exposition HTTP sont strictement réservés au développement local Docker Desktop.

## Backend Go

Le backend Go, son endpoint utilisateur OIDC et son chart PostgreSQL se trouvent dans [`backend`](./backend) et [`deploy/helm/kubecoder`](./deploy/helm/kubecoder). Le guide du chart décrit le build de l'image locale et l'installation dans Kubernetes.
