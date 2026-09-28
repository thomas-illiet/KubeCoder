# Stratégie de validation

## Pyramide de tests

- **Unitaires** : résolution des scopes, machine d'états, autorisation, snapshots canoniques, redaction.
- **Composants** : repositories PostgreSQL, fournisseur de secrets simulé, client Kubernetes fake, adaptateur runtime simulé.
- **MCP** : résolution des portées et bindings, adaptateurs STDIO/SSE, worker Docker, timeout et nettoyage idempotent.
- **Contrats** : OpenAPI, `RunEvent v1`, adaptateur moteur et compatibilité des snapshots.
- **Intégration** : PostgreSQL réel, stockage objet, IdP OIDC de référence et cluster kind/k3d.
- **E2E navigateur** : Playwright sur build de production Vue/Vuetify.
- **Sécurité** : accès inter-tenant, CSRF, sessions, images, Pod Security, réseau et secret scanning.
- **Résilience** : suppression Pod, restart API/reconciler, événements dupliqués, coffre indisponible, stockage plein.
- **Charge** : file de runs, fan-out SSE, débit d'événements et ordonnancement équitable.

## Scénarios d'acceptation transverses

1. Deux utilisateurs de deux organisations créent des ressources aux mêmes noms sans visibilité croisée.
2. Un membre non administrateur ne peut ni créer ni résoudre un secret par API directe.
3. Une rotation de secret n'altère pas un run existant et s'applique au suivant.
4. Un skill désactivé au repository n'apparaît pas dans le snapshot effectif, même s'il est globalement actif.
5. Le digest affiché avant lancement égale celui enregistré sur le run.
6. Deux requêtes portant la même clé d'idempotence ne créent qu'un run.
7. Après coupure SSE, `Last-Event-ID` restitue les événements manquants sans doublon visible.
8. Après restart du reconciler, les runs repartent de l'état canonique sans créer deux Pods.
9. Une annulation supprime le workload et le Secret temporaire, tout en conservant le chat déjà persisté.
10. Un secret volontairement écrit par le programme agent est bloqué/redacté selon la politique et produit une alerte de sécurité.
11. Une URL d'artefact expirée ou utilisée par une autre organisation est refusée.
12. Une reprise indique honnêtement `NATIVE`, `RECONSTRUCTED` ou `FRESH`.
13. Un membre peut choisir un agent publié et configurer les options autorisées de son repository, mais les endpoints d'administration refusent toute lecture ou mutation de credentials et paramètres sensibles.
14. La projection utilisateur d'un agent ne contient ni prompt système, ni référence de coffre, ni credential, ni paramètre classé sensible.
15. Un secret émis en plusieurs chunks et dans une variante supportée est remplacé avant persistance et SSE ; les tables, objets et logs ne contiennent jamais la valeur brute.
16. Une sortie suspecte non classifiable est mise en quarantaine et n'atteint pas le navigateur.
17. Un secret expiré reste visible dans l'historique mais bloque un nouveau run qui tente de le résoudre, avec une erreur actionnable ne révélant aucune valeur.
18. Un serveur MCP global apparaît désactivé dans une organisation autorisée et ne devient utilisable qu'après activation explicite ; un serveur MCP d'organisation reste invisible à tout autre tenant.
19. Un test STDIO n'exécute qu'un binaire et des arguments approuvés ; un test SSE refuse loopback, metadata cloud, redirection ou résolution hors allowlist.
20. Chaque test utilise un conteneur Docker dédié puis prouve la suppression du conteneur, du réseau, des volumes et des secrets temporaires, y compris après timeout ou crash du worker.
21. Les snapshots, API, événements et logs ne contiennent aucune valeur de header, variable ou secret MCP ; le résultat de test est borné et expurgé.
22. Un agent ne peut ni découvrir ni appeler un serveur ou un outil MCP absent de ses bindings figés dans le snapshot du run.

## Gates CI/CD

- format, lint, tests unitaires et détection de secrets ;
- génération OpenAPI sans diff non commité ;
- migrations montante et rollback lorsque supporté ;
- build frontend production et images OCI reproductibles ;
- SBOM, scan de vulnérabilités et signature d'image ;
- tests d'intégration puis E2E sur environnement éphémère ;
- policy checks Kubernetes ;
- déploiement préproduction, smoke tests, puis promotion du même digest.

## Definition of Done d'une fonctionnalité

- comportement heureux et refus d'autorisation couverts ;
- erreur utilisateur actionnable ;
- événement d'audit si sensible ;
- métrique/log corrélé sans secret ;
- idempotence et reprise définies ;
- documentation/migration ajoutées ;
- validation visuelle desktop et mobile si l'UI change ;
- preuve d'exécution conservée dans la CI.

