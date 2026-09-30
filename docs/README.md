# Plan directeur — KubeCoder

Ce dossier décrit le produit, son architecture cible et l'ordre recommandé de réalisation. Il s'agit d'un plan vivant : les choix encore à arbitrer sont signalés comme tels et centralisés dans [`12-questions-ouvertes.md`](./12-questions-ouvertes.md).

## Objectif

KubeCoder permet à un utilisateur authentifié de démarrer, suivre, interrompre puis reprendre une session d'agent de code exécutée à la demande dans Kubernetes. Chaque session est rattachée à une organisation et à un repository, possède une configuration d'agent résolue et immuable par run, et est représentée par un chat persistant dans l'interface web.

Le moteur d'agent retenu est OpenCode. Son dépôt/distribution, sa version et son contrat de reprise exacts restent à confirmer avant l'implémentation de l'adaptateur d'exécution.

## Lecture recommandée

| Fichier | Contenu |
|---|---|
| [`00-vision-et-perimetre.md`](./00-vision-et-perimetre.md) | Vision, acteurs, MVP et hors périmètre |
| [`01-modele-de-domaine.md`](./01-modele-de-domaine.md) | Vocabulaire et entités métier |
| [`02-architecture-cible.md`](./02-architecture-cible.md) | Architecture logique et composants |
| [`03-cycle-execution-kubernetes.md`](./03-cycle-execution-kubernetes.md) | Cycle de vie d'un run et reprise |
| [`04-organisations-et-autorisations.md`](./04-organisations-et-autorisations.md) | Isolation, rôles et règles d'accès |
| [`05-secrets.md`](./05-secrets.md) | Stockage, résolution et injection des secrets |
| [`06-skills-et-configuration-agent.md`](./06-skills-et-configuration-agent.md) | Héritage des skills et snapshot de configuration |
| [`07-chat-et-sessions.md`](./07-chat-et-sessions.md) | Modèle de chat, événements et continuité |
| [`08-interface-vuetify.md`](./08-interface-vuetify.md) | Parcours et direction visuelle sombre |
| [`09-securite-et-exploitabilite.md`](./09-securite-et-exploitabilite.md) | Défense en profondeur, audit et observabilité |
| [`10-plan-execution.md`](./10-plan-execution.md) | Lots, dépendances et critères de sortie |
| [`11-strategie-de-validation.md`](./11-strategie-de-validation.md) | Tests et critères d'acceptation globaux |
| [`12-questions-ouvertes.md`](./12-questions-ouvertes.md) | Arbitrages attendus, classés par priorité |
| [`13-journal-decisions.md`](./13-journal-decisions.md) | Décisions proposées, acceptées ou remplacées |
| [`14-serveurs-mcp.md`](./14-serveurs-mcp.md) | Catalogue MCP global/organisation, transports STDIO/SSE et test Docker isolé |

## Principes directeurs

- Le backend est le point de contrôle : le navigateur ne reçoit ni jeton OIDC fournisseur, ni secret de repository, ni accès Kubernetes.
- Les définitions d'agents, leurs credentials et leurs paramètres sensibles sont administrés hors des parcours utilisateur. Un utilisateur choisit uniquement parmi les agents qui lui sont publiés et configure les paramètres non sensibles de son repository.
- PostgreSQL est la source de vérité du plan de contrôle. Kubernetes est réconcilié à partir d'un état désiré, jamais piloté par des commandes `kubectl` lancées depuis une requête HTTP.
- Une `Session` représente la conversation durable. Un `Run` représente une tentative Kubernetes bornée et remplaçable.
- La configuration effective d'un run est enregistrée avant démarrage avec son empreinte ; une modification de skill, de serveur MCP ou de secret ne change pas rétroactivement un run déjà créé.
- Les frontières d'organisation sont appliquées dans les requêtes, l'autorisation, les identifiants Kubernetes, le stockage d'objets et l'audit.
- Keycloak sert d'IdP de référence, mais l'application ne dépend que des standards OIDC/OAuth.
- Un événement susceptible de contenir un secret est masqué avant persistance et avant diffusion ; le navigateur ne reçoit pas la valeur brute.

## Statuts de décision

- **Proposé** : base de travail à confirmer.
- **Accepté** : validé explicitement et utilisable pour implémenter.
- **Bloquant** : doit être tranché avant le lot indiqué.
- **Différé** : volontairement hors du premier incrément.

