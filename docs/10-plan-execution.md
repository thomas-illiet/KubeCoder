# Plan d'exécution global

Les durées ne sont volontairement pas chiffrées avant confirmation de l'équipe, du moteur exact et de l'environnement Kubernetes. Chaque lot se termine par une démonstration et des preuves automatisées.

## Lot 0 — Décisions et preuve du moteur

**But** : lever les inconnues qui pourraient invalider l'architecture.

- Confirmer la distribution/version d'OpenCode, son image, sa licence, sa CLI/API et son protocole d'événements.
- Construire un spike local puis Kubernetes : message entrant, événements, arrêt et reprise.
- Confirmer cluster, ingress, StorageClass, registre, DNS, coffre et stockage objet.
- Choisir stratégie namespace, workspace et rétention.
- Écrire les ADR acceptés et le contrat `AgentRuntimeAdapter v1`.

**Sortie** : un Pod de test exécute réellement le moteur, exporte des événements normalisés et documente précisément la capacité de reprise.

## Lot 1 — Socle du plan de contrôle

**But** : disposer d'un produit déployable, authentifié et observable.

- Monorepo ou dépôts, CI, conventions, génération OpenAPI et migrations.
- API Go, frontend Vue/Vuetify sombre fondé sur des composants génériques et réutilisables, PostgreSQL et health/readiness.
- BFF OIDC standard et Keycloak de référence pour le développement.
- Utilisateurs, organisations, memberships et sélecteur d'organisation.
- Audit de base et corrélation des requêtes.

**Dépend de** : décisions OIDC et cluster du lot 0.

**Sortie** : connexion, changement d'organisation et preuve de refus d'accès croisé.

## Lot 2 — Repositories et configuration

**But** : décrire de façon sûre ce que l'agent devra exécuter.

- CRUD repository, validation Git et credentials référencés.
- Définitions d'agents administrées, versionnées et publiées, avec projection utilisateur expurgée.
- Association repository-agent et schéma borné des options configurables par l'utilisateur.
- Secrets organisation/repository, rotation et bindings explicites.
- Skills global/organisation/repository, tri-state et vue de statut effectif.
- Catalogue MCP global/organisation, bindings explicites aux agents, transports STDIO/SSE et projections expurgées.
- Worker de test séparé lançant un conteneur Docker éphémère dédié avec handshake, `tools/list`, résultat nettoyé et nettoyage durable.
- Génération et validation du snapshot canonique avec digest.

**Dépend de** : lot 1 ; contrat adaptateur du lot 0.

**Sortie** : prévisualisation exacte d'une configuration exécutable, sans valeur secrète dans l'API.

## Lot 3 — Orchestration Kubernetes verticale

**But** : lancer un premier run complet depuis l'interface.

- Tables run/outbox et worker de réconciliation idempotent.
- Provisioning Pod/Job, init repository, workspace, configuration et Secret temporaire.
- Bridge/adaptateur du moteur, jeton court et ingestion d'événements.
- Provisioning des MCP STDIO liés au run et client SSE limité aux endpoints et outils figés dans le snapshot.
- Redaction streaming dans le bridge puis l'ingestor avant toute persistance ou diffusion.
- États, annulation, timeout et nettoyage durable.
- Sécurité Pod et NetworkPolicies minimales.

**Dépend de** : lots 0 et 2.

**Sortie** : un message web déclenche un run réel et retourne un résultat durable.

## Lot 4 — Chat persistant et reprise

**But** : transformer l'exécution ponctuelle en session continue.

- Modèle session/message/event/artifact.
- SSE reprenable, projection des messages et UX de reconnexion.
- Plusieurs runs séquentiels dans une session.
- Checkpoint natif ou reconstruction déclarée.
- PVC/archive workspace, TTL et restauration.
- Vue fichiers/artefacts et export de session non secret.

**Dépend de** : lot 3 et résultat du spike de reprise.

**Sortie** : fermeture/réouverture du navigateur et reprise après remplacement de Pod démontrées en E2E.

## Lot 5 — Durcissement multi-tenant

**But** : rendre le service acceptable pour une bêta contrôlée.

- Tests systématiques inter-organisations et revue du schéma d'autorisation.
- Sandbox runtime, politiques réseau complètes et supply-chain OCI.
- Scénarios de fuite/redaction des secrets.
- Fair scheduling, backpressure et protection anti-abus.
- Sauvegarde/restauration et exercices de panne.
- Tableaux de bord, alertes et runbooks.

**Dépend de** : lots 1 à 4.

**Sortie** : revue de sécurité sans défaut critique connu et exercice de restauration réussi.

## Lot 6 — Bêta puis production

**But** : exploiter avec un groupe limité puis élargir.

- SLO, RPO/RTO, capacité et tests de charge.
- Migrations/rollbacks, canary et compatibilité des snapshots/adaptateurs.
- Politiques de rétention et suppression vérifiable.
- Documentation opérateur/utilisateur et support.
- Revue des retours bêta puis critères go/no-go.

**Sortie** : SLO mesurés, procédures testées et approbation explicite de mise en production.

## Chemin critique

1. Identifier le moteur et prouver son protocole/reprise.
2. Fixer l'isolation Kubernetes et la persistance du workspace.
3. Stabiliser snapshot + adaptateur + événements.
4. Construire l'orchestration verticale avant d'élargir l'interface d'administration.
5. Prouver la continuité et l'absence d'accès croisé.

## Règles transverses pour le frontend Vue/TypeScript

- Tout élément d'interface récurrent doit être extrait en composant Vue générique, typé et réutilisable au lieu d'être réimplémenté dans chaque écran. Cela couvre notamment les modales, data tables, formulaires, champs de recherche, filtres, états de chargement, états vides, confirmations et affichages d'erreur.
- Les composants génériques doivent exposer des contrats TypeScript explicites (props, emits, slots et types génériques lorsque nécessaire) et rester configurables sans embarquer de logique métier propre à un écran. Les composants métier les composent et leur fournissent les colonnes, actions, données et contenus spécifiques.
- Un composant partagé de modal doit centraliser au minimum le titre, les actions, la fermeture, le comportement responsive, le focus clavier, l'accessibilité et les états de chargement ou d'erreur. Son panneau utilise toujours un fond opaque contrasté et un scrim suffisant pour que le contenu sous-jacent ne nuise pas à la lecture.
- Un composant partagé de data table doit centraliser au minimum la pagination, le tri, le chargement, l'état vide, les erreurs, les actions de ligne et les slots de rendu, tout en conservant un typage strict des lignes et des colonnes.
- Les modèles et schémas de validation doivent eux aussi être mutualisés, fortement typés et composables. Une règle commune ne doit pas être dupliquée dans plusieurs formulaires ; les validations métier propres à un écran étendent les briques partagées.
- Les pages de recherche et de liste doivent dissocier visuellement les responsabilités : les champs de recherche et filtres sont placés dans une `v-card` dédiée, tandis que le tableau et ses actions appartiennent à une autre `v-card`. Les en-têtes, résumés ou actions globales peuvent utiliser des cartes distinctes lorsque cela améliore la lecture.
- Cette séparation en cartes doit rester cohérente sur desktop, tablette et mobile, avec des espacements explicites et sans imbriquer le tableau dans la carte des filtres.
- L'interface doit être conçue en priorité pour l'usage quotidien d'un développeur : vocabulaire familier, informations techniques utiles au bon moment, raccourcis clavier pertinents, valeurs par défaut sûres et actions fréquentes accessibles sans navigation inutile.
- Le parcours principal, depuis le choix d'une organisation et d'un repository jusqu'au démarrage puis à la reprise d'un chat, doit être immédiatement compréhensible. Il doit guider l'utilisateur étape par étape, rendre visibles le repository, la branche, le profil d'agent et les skills effectifs, puis présenter un résumé vérifiable avant toute exécution.
- Chaque écran doit avoir une action principale identifiable, une hiérarchie visuelle claire et une divulgation progressive des options avancées. Les détails rares ou complexes ne doivent pas surcharger le parcours nominal, mais doivent rester accessibles sans impasse.
- Toute action asynchrone doit fournir un retour immédiat et explicite : chargement, progression, succès, erreur exploitable et possibilité de réessayer. Les libellés d'erreur doivent indiquer ce qui s'est produit et la prochaine action possible, sans exposer de secret.
- Le contexte de travail doit être conservé autant que possible : organisation, repository, branche, filtres, brouillon du message et session courante ne doivent pas être perdus lors d'une navigation, d'une reconnexion ou d'une erreur récupérable.
- Les parcours frontend critiques doivent faire l'objet de tests d'utilisabilité et de tests E2E couvrant au minimum la création d'une session, le premier message envoyé au repository, la compréhension de l'état du run, la reprise du chat et la résolution d'une erreur courante.
- Chaque composant partagé doit disposer de tests ciblant son contrat public et ses principaux états. Chaque écran consommateur doit prouver par un test d'intégration qu'il compose correctement les composants génériques, y compris la séparation entre carte de recherche et carte de résultats.

## Organisation du backlog

Chaque story doit indiquer : organisation concernée, permission exigée, impact secret, état d'idempotence, métriques, migration, scénario de panne et preuve d'acceptation. Toute story frontend doit également identifier les composants génériques créés ou réutilisés, les modèles de validation partagés, la composition des `v-card`, l'action principale, les états de retour utilisateur et l'impact sur le parcours développeur. Une fonctionnalité touchant l'exécution n'est pas terminée sans test sur un cluster éphémère ou de préproduction.

