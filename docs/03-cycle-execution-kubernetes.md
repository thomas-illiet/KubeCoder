# Cycle d'exécution Kubernetes

## Création atomique d'un run

La requête `POST /sessions/{id}/runs` doit :

1. vérifier l'appartenance, le rôle et l'absence d'un run concurrent ;
2. valider le message, l'association repository-agent et la version publiée de l'agent ;
3. calculer la configuration effective sans lire les valeurs de secrets ;
4. écrire message, snapshot, run et outbox dans une même transaction ;
5. retourner immédiatement le run et son état `QUEUED`.

Une clé d'idempotence client évite qu'un double clic crée deux runs.

## Réconciliation

Le worker traite chaque état de façon idempotente :

- `QUEUED` : attend une capacité d'exécution disponible.
- `RESOLVING` : vérifie versions d'image, références de skills, secrets et repository.
- `PROVISIONING` : crée Secret/ConfigMap temporaires, PVC si nécessaire et Job/Pod.
- `STARTING` : attend la disponibilité du bridge et publie une progression compréhensible.
- `RUNNING` : renouvelle le lease, reçoit les événements et surveille timeout/annulation.
- terminal : collecte résultat, révoque le jeton, supprime les secrets Kubernetes et programme le nettoyage.

Chaque ressource Kubernetes reçoit labels et annotations `organization_id`, `session_id`, `run_id`, `snapshot_digest` et `managed-by=kubecoder`, sans données sensibles.

## Workspace et reprise

Proposition pour le MVP : un workspace durable par session, monté sur un PVC `ReadWriteOnce`, avec un seul run actif. Le repository est cloné au premier run puis réutilisé. Avant reprise :

- vérifier que le remote et la branche n'ont pas changé de manière incompatible ;
- restaurer l'état natif de l'agent si le moteur le supporte ;
- reconstruire le contexte depuis les messages/snapshots si l'état natif est absent ;
- afficher clairement à l'utilisateur si la reprise est exacte ou reconstruite.

Une politique de TTL archive puis supprime les workspaces inactifs. Les métadonnées de session survivent au workspace selon leur propre rétention.

## Annulation, timeout et panne

- **Annulation utilisateur** : état `CANCELING`, signal gracieux, délai, puis suppression forcée du Pod.
- **Timeout** : limite par run ; événement `TIMED_OUT` et conservation des sorties déjà persistées.
- **Pod supprimé** : le reconciler classe le run `ORPHANED` puis `RECOVERY_REQUIRED` ou `FAILED` selon la politique.
- **API redémarrée** : aucun run n'est perdu ; le reconciler reprend depuis PostgreSQL et l'état observé du cluster.
- **Bridge déconnecté** : reprise avec dernière séquence acquittée ; doublons dédupliqués.
- **Échec de nettoyage** : tâche durable avec retry et alerte, sans masquer le résultat fonctionnel du run.

## Concurrence et capacité

- Verrou applicatif/SQL sur la session lors de la création d'un run.
- File équitable entre organisations afin qu'une seule organisation ne monopolise pas le cluster.
- Bornes techniques non configurables dans le produit : `LimitRange`, `activeDeadlineSeconds` et TTL du Job.

## Preuve de continuité attendue

Un test de bout en bout doit démarrer un run, recevoir des événements, fermer le navigateur, supprimer le Pod ou interrompre le réseau, rouvrir la session et démontrer soit une reprise, soit un état d'échec explicite suivi d'un nouveau run continuant le contexte.

