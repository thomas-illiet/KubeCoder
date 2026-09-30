# Journal des décisions

Ce fichier sert de registre léger. Une décision structurante acceptée pourra ensuite devenir un ADR séparé dans `Project/decisions/`.

| ID | Sujet | Statut | Décision / proposition | Conséquence |
|---|---|---|---|---|
| D-001 | Frontend | Proposé | Vue 3 + TypeScript + Vuetify 3, thème sombre | Conforme au besoin UI, à confirmer avant scaffold |
| D-002 | Auth navigateur | Proposé | Backend Go comme OIDC RP/BFF, jetons serveur, cookie opaque | Portable entre Keycloak et tout IdP OIDC conforme |
| D-003 | Autorisation | Proposé | Rôles locaux `OWNER`, `ADMIN`, `MEMBER` | Aucun claim Keycloak spécifique requis |
| D-004 | Source de vérité | Proposé | PostgreSQL pour l'état du plan de contrôle | Kubernetes reste un état observé/réconcilié |
| D-005 | Temps réel | Proposé | HTTP commandes + SSE reprenable | WebSocket seulement si exigé par le runtime |
| D-006 | Unité de continuité | Proposé | Session durable, plusieurs runs Kubernetes séquentiels | Un Pod n'est pas une session |
| D-007 | Configuration | Proposé | Snapshot immuable par run avec digest | Reproductibilité et audit |
| D-008 | Skills | Proposé | Global -> organisation -> repository -> session autorisée | Override spécifique et statut effectif explicite |
| D-009 | Secrets | Proposé | Bindings explicites, valeurs hors tables métier | Évite l'injection implicite de tous les secrets visibles |
| D-010 | Workspace | Proposé | PVC par session et un run actif au MVP | Reprise simple, coût/TTL à confirmer |
| D-011 | Isolation Kubernetes | Bloquant | Namespace d'exécution par organisation recommandé | Dépend du niveau de menace et des capacités cluster |
| D-012 | Runtime agent | Accepté | OpenCode est encapsulé derrière `AgentRuntimeAdapter v1` | Évite de coupler le domaine à sa CLI ou son protocole natif |
| D-013 | Eventing | Proposé | Outbox PostgreSQL avant ajout d'un broker | Réduit la complexité initiale |
| D-014 | Secret provider | Bloquant | Vault/KMS préféré, enveloppe applicative en repli | Dépend de l'infrastructure disponible |
| D-015 | Administration des agents | Accepté | Les administrateurs configurent/publient agents, credentials et paramètres sensibles ; l'utilisateur choisit un agent autorisé et configure seulement les options repository exposées | Séparation stricte entre définition privée et projection utilisateur |
| D-016 | Masquage conversation | Accepté | Double redaction bridge + ingestor avant persistance et diffusion, avec quarantaine en cas d'ambiguïté | Le navigateur et le stockage courant ne reçoivent pas la valeur brute |
| D-017 | Distribution OpenCode | Bloquant | Dépôt, version, image, licence et contrat de reprise à confirmer | Nécessaire avant le spike runtime du lot 0 |
| D-018 | Limites et quotas produit | Accepté | Ne pas exposer de fonctionnalité d'administration des limites ou quotas ; conserver seulement les bornes techniques nécessaires à la sécurité des workloads | Aucun écran ni réglage de quota dans le produit |
| D-019 | Configuration LLM des agents | Accepté | Chaque définition d'agent configure son endpoint OpenAI ou compatible et sa clé API en écriture seule ou par référence de credential | Endpoint privé et clé absents de la projection utilisateur |
| D-020 | Expiration des secrets | Accepté | Une version de secret peut expirer et expose un état calculé ; une version expirée ne peut plus être résolue pour un nouveau run | Alertes préalables, historique conservé et erreur de run actionnable |
| D-021 | Séparation organisation / administration | Accepté | Deux layouts de routage distincts : Organisation comme entrée par défaut et Administration ouverte depuis une entrée dédiée en pied de menu | Chaque layout possède sa propre navigation ; l’Admin représente le service global et n’affiche ni organisation active, ni sélecteur, ni breadcrumb tenant |
| D-022 | Serveurs MCP | Accepté | Catalogue MCP aux scopes global et organisation ; transports MVP limités à STDIO et SSE ; aucun état d'activation dans l'administration et tous les bindings organisation désactivés par défaut | L'organisation peut activer un MCP global mais ne peut pas modifier sa définition ; aucun MCP n'est présélectionné dans un agent et le snapshot fige versions, transports et outils autorisés |
| D-023 | Test des serveurs MCP | Accepté | Chaque test lance un conteneur Docker éphémère dédié depuis un worker isolé, exécute handshake et `tools/list`, puis garantit le nettoyage | L'API n'accède pas à la socket Docker ; réseau, ressources, durée, logs et secrets sont strictement bornés |
| D-024 | Gestion des secrets d'organisation | Accepté | Le layout organisation expose aux `OWNER` et `ADMIN` la liste, la création et la rotation des secrets organisation/repository, toujours en écriture seule | Les secrets globaux de plateforme sont absents des API et vues organisation ; les membres standard n'y accèdent pas et l'administration globale ne voit jamais les valeurs tenant |
| D-025 | Accès Git global | Accepté | L'accès Git utilise une clé publique unique configurée globalement dans le backend ; aucun credential Git n'est géré depuis la page Secrets | La page d'administration Secrets ne liste et ne crée que des secrets applicatifs ; la configuration de la clé Git reste hors de l'interface |
| D-026 | Authentification du prototype local | Accepté | Le prototype Vue utilise temporairement un client OIDC public générique, Authorization Code + PKCE S256 et des tokens en `sessionStorage`, contre Keycloak local ; aucun adaptateur `keycloak-js` ni rôle applicatif | Toutes les routes métier exigent une session. Le BFF Go et son cookie opaque restent l'architecture cible avant toute mise en production |

## Modèle d'une décision

```md
## D-XXX — Titre

- Statut : proposé | accepté | remplacé | abandonné
- Date : YYYY-MM-DD
- Contexte : pourquoi la décision est nécessaire
- Décision : choix retenu
- Alternatives : options écartées et raisons
- Conséquences : effets positifs, coûts et risques
- Validation : preuve ou test attendu
```
