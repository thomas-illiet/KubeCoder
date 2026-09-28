# Kube Coder

Le plan directeur du projet se trouve dans [`Project/README.md`](./Project/README.md).

## Prototype du panneau d’administration

Le prototype est une application frontend en Vue 3, TypeScript et Vuetify 3. L’authentification, le profil, les organisations et leurs memberships utilisent le backend réel. Les autres écrans métier contiennent encore des données fictives.

```sh
helm upgrade --install kubecoder-auth ./deploy/helm/kubecoder-keycloak \
  --namespace kubecoder-auth --create-namespace --wait --timeout 10m

cd frontend
npm install
npm run dev
```

Le frontend charge sa configuration à l'exécution depuis
[`frontend/public/config.json`](./frontend/public/config.json). Ce fichier définit
l'URL du backend ainsi que les paramètres OIDC et doit être remplacé par la
configuration propre à chaque environnement sans reconstruire l'application.

```json
{
  "apiBaseUrl": "http://localhost:8080",
  "oidc": {
    "authority": "http://localhost:30080/realms/kubecoder",
    "clientId": "kubecoder-web",
    "redirectUri": "http://localhost:5173/auth/callback",
    "postLogoutRedirectUri": "http://localhost:5173/logout/callback"
  }
}
```

Toutes les propriétés sont obligatoires et les URL doivent être absolues. Le
fichier est chargé avec `cache: no-store` avant l'initialisation OIDC ; une
configuration absente ou invalide bloque volontairement le démarrage.

L'interface est alors disponible sur `http://localhost:5173/organization`. Sans membership, elle présente un état vide ; un administrateur plateforme peut créer les organisations et gérer leurs membres dans `/admin/organizations`.

Le compte local est `admin` / `admin`. Le frontend utilise une bibliothèque OIDC générique, le flux Authorization Code avec PKCE S256 et `sessionStorage`; il ne dépend pas de `keycloak-js`. Les routes `/organizations/:slug` exigent un membership et les routes `/admin` exigent le droit administrateur retourné par le backend.

Pour vérifier le build de production :

```powershell
cd frontend
npm run build
```

Les identifiants présents dans le chart et l’exposition HTTP sont strictement réservés au développement local Docker Desktop.

## Backend Go

Le backend Go, son endpoint utilisateur OIDC et son chart PostgreSQL se trouvent dans [`backend`](./backend) et [`deploy/helm/kubecoder`](./deploy/helm/kubecoder). Le guide du chart décrit le build de l'image locale et l'installation dans Kubernetes.
