# Organisations et autorisations

## Modèle d'isolation

L'organisation est la frontière principale de propriété. Un utilisateur peut appartenir à plusieurs organisations et doit toujours voir l'organisation courante dans la navigation. L'URL porte un identifiant lisible ou un slug, mais l'autorisation utilise l'identifiant interne.

Le sélecteur doit rester utilisable avec plusieurs milliers d'organisations. Il affiche d'abord un petit nombre d'organisations récentes, puis utilise une recherche côté serveur, temporisée et paginée par nom ou slug. Le navigateur ne charge jamais la liste complète et le changement conserve une indication explicite de l'organisation sélectionnée.

Règles minimales :

- aucune ressource tenant-aware n'est accessible sans membership actif ;
- toute requête est évaluée avec `user_id + organization_id + action + resource` ;
- changer d'organisation invalide les caches frontend tenant-aware ;
- les tâches asynchrones transportent explicitement `organization_id` et le vérifient à nouveau ;
- les données d'une organisation ne sont jamais recherchées par un identifiant de ressource seul.

## Rôles proposés

| Action | OWNER | ADMIN | MEMBER |
|---|:---:|:---:|:---:|
| Voir repositories, sessions, agents publiés et skills effectifs | Oui | Oui | Oui |
| Créer une session et lancer un run | Oui | Oui | Oui, selon la politique de l'organisation |
| Ajouter/configurer les champs non sensibles d'un repository | Oui | Oui | Oui, sur les repositories autorisés |
| Choisir un agent publié pour un repository | Oui | Oui | Oui, sur les repositories autorisés |
| Créer/modifier une définition d'agent ou ses paramètres sensibles | Selon scope plateforme | Oui, pour les agents d'organisation si autorisé | Non |
| Affecter credentials et bindings de secrets | Oui | Oui | Non |
| Créer ou changer un secret | Oui | Oui | Non |
| Voir une valeur de secret après création | Non | Non | Non |
| Gérer skills et agents d'organisation | Oui | Oui | Non |
| Inviter/retirer des membres | Oui | Oui, sauf owners | Non |
| Changer les rôles owner | Oui | Non | Non |
| Supprimer/transférer l'organisation | Oui | Non | Non |

Le rôle `PLATFORM_ADMIN` est global et séparé des memberships. Son usage doit être audité et ne doit pas lui donner implicitement accès au contenu des secrets.

## Vue utilisateur d'un agent

L'API utilisateur retourne une projection dédiée : identifiant, nom, description, capacités, compatibilité, ressources visibles et schéma des options repository autorisées. Elle exclut provider/model internes si classés sensibles, prompt système, références de credentials, noms de coffre, variables protégées et politiques réseau internes. La sélection se fait par identifiant opaque ; elle ne permet pas de reconstruire les secrets ni la configuration d'administration.

## Authentification OIDC portable

- Authorization Code Flow via le backend ; PKCE et vérification `state`, `nonce`, signature, issuer et audience.
- Discovery via `/.well-known/openid-configuration`.
- Jetons fournisseur stockés côté serveur ; cookie navigateur opaque `HttpOnly`, `Secure`, `SameSite=Lax` ou plus strict selon les flux.
- Identité stable basée sur `(issuer, sub)`, jamais sur l'adresse e-mail seule.
- Mapping configurable des claims d'affichage ; les rôles applicatifs restent en base locale.
- Rotation des clés de signature, expiration/refresh, logout local et logout fournisseur lorsque supporté par le standard.
- Keycloak fournit la configuration et les tests de référence, sans SDK, API d'administration, route ou claim spécifique dans le domaine.

## Défense côté données

- Repositories et services exigent `organization_id` dans leur interface.
- Tests systématiques d'accès croisé entre deux organisations.
- PostgreSQL Row-Level Security peut renforcer l'isolation, mais ne remplace pas l'autorisation applicative ; décision à prendre avant le schéma final.
- Les exports, URLs signées et événements SSE portent la même vérification d'appartenance.

