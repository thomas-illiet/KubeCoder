# Vision et périmètre

## Proposition de valeur

Un membre d'organisation sélectionne un repository et un agent préconfiguré par un administrateur, associe cet agent à la configuration non sensible de son repository, ouvre une session de chat et envoie une demande. KubeCoder prépare un environnement isolé dans Kubernetes, injecte uniquement les secrets autorisés, résout les skills applicables, exécute le moteur d'agent et retransmet des événements préalablement nettoyés. La conversation, les décisions d'exécution et les artefacts restent disponibles afin de reprendre le travail plus tard.

## Acteurs

- **Administrateur plateforme** : crée et publie les agents, gère leurs paramètres sensibles et credentials, les skills globaux et l'exploitation.
- **Owner d'organisation** : gère l'organisation, ses membres, ses repositories et ses secrets.
- **Admin d'organisation** : administre les ressources sans pouvoir supprimer/transférer l'organisation, selon la politique retenue.
- **Member** : configure les propriétés non sensibles de ses repositories, choisit un agent publié et utilise les sessions auxquelles l'organisation lui donne accès.
- **Contrôleur d'exécution** : réconcilie les runs demandés avec les workloads Kubernetes.
- **Moteur d'agent** : processus OpenCode encapsulé derrière un adaptateur versionné.

## Parcours MVP

1. Se connecter via un fournisseur OIDC compatible, Keycloak servant de référence.
2. Choisir une organisation ou en créer une si la politique l'autorise.
3. Enregistrer/configurer un repository Git et choisir un agent dans le catalogue autorisé, sans voir les credentials associés.
4. Un administrateur crée les secrets d'organisation ou de repository et leurs bindings de variables d'environnement.
5. L'utilisateur consulte les skills effectifs ; seuls les administrateurs modifient les politiques qui ne lui sont pas déléguées.
6. Créer une session liée au repository et à son association d'agent.
7. Envoyer un message et voir le run passer par les états de préparation, exécution et fin.
8. Suivre en direct les messages de l'agent, demandes d'action, erreurs et artefacts.
9. Fermer la page, revenir à la session puis continuer avec un nouveau run.
10. Arrêter ou annuler une exécution et consulter son audit.

## Inclus dans le MVP

- OIDC standard avec session navigateur opaque.
- Organisations, membres et rôles locaux.
- Repositories Git et état de connectivité.
- Secrets aux scopes organisation et repository.
- Skills globaux, organisation et repository, avec calcul du statut effectif.
- Configurations d'agent versionnées.
- Sessions de chat persistantes et plusieurs runs par session.
- Exécution Kubernetes à la demande, événements en direct et reprise.
- Thème sombre Vuetify, responsive desktop/tablette/mobile.
- Audit des opérations sensibles et métriques d'exploitation.

## Hors MVP, sauf besoin confirmé

- Facturation.
- Marketplace publique de skills.
- Édition collaborative temps réel d'une même session.
- Exécution hors Kubernetes.
- Fédération multi-cluster et placement géographique.
- Hooks GitHub/Jira automatiques.
- Exposition d'une API publique avec jetons Bearer utilisateur.
- IDE web complet ; le MVP expose chat, fichiers modifiés et artefacts ciblés.

## Indicateurs de réussite initiaux

- Un utilisateur peut créer une session et recevoir le premier événement d'agent sans accès direct au cluster.
- Une session interrompue par fermeture du navigateur peut être rouverte sans perte de messages.
- La suppression d'un Pod en cours produit un état explicite et permet une reprise contrôlée.
- Aucun secret en clair n'apparaît dans les réponses API, journaux applicatifs ou événements du chat.
- Une valeur secrète volontairement affichée par OpenCode est remplacée avant d'atteindre le navigateur et n'est pas persistée en clair.
- Deux organisations ne peuvent pas lire, utiliser ou lister leurs ressources respectives.
- La configuration effective affichée avant lancement correspond au snapshot utilisé par le run.

