# Questions ouvertes

Les réponses doivent être reportées dans le journal de décisions. Les questions P0 bloquent une partie du chemin critique.

## P0 — À trancher avant l'implémentation du runtime

1. **Quelle distribution d'OpenCode fait autorité ?** Quel dépôt, version, image OCI, licence et documentation devons-nous utiliser ?
2. **Quel est son contrat réel ?** CLI, serveur, SDK, format d'événements, gestion des outils, interruption et export/import d'une session ?
3. **Quel Kubernetes cible-t-on ?** Distribution/version, cluster existant ou à créer, Ingress, StorageClass, registre OCI, stockage objet et politique d'admission disponibles ?
4. **Quel niveau d'isolation est attendu ?** Utilisateurs internes de confiance, code potentiellement hostile, ou service multi-tenant exposé ? Ce choix conditionne namespace par organisation, gVisor/Kata et egress.
5. **Que signifie “continuer l'exécution” ?** Reprendre exactement la session native de l'agent, conserver seulement le chat, conserver aussi le workspace Git, ou les trois ?

## P0 — À trancher avant le schéma secrets/repositories

6. Quels fournisseurs Git faut-il supporter au MVP : URL générique SSH/HTTPS, GitHub, GitLab, Bitbucket, GitHub Enterprise/Data Center ?
7. Les repositories sont-ils seulement clonés ou l'agent peut-il pousser des commits/branches et ouvrir des pull requests ?
8. Quel coffre est disponible : Vault, cloud secret manager, SOPS/KMS, aucun ?
9. Un secret d'organisation est-il disponible à tous ses repositories ou seulement après binding explicite ? Le plan recommande le binding explicite.
10. Faut-il autoriser un secret repository à masquer une variable d'organisation du même nom, ou bloquer toute collision ?

## P1 — Produit et gouvernance

11. Qui peut créer une organisation et inviter des membres ?
12. Les rôles `OWNER`, `ADMIN`, `MEMBER` suffisent-ils ? Faut-il des permissions par repository ?
13. Les sessions sont-elles privées à leur auteur ou visibles par tous les membres autorisés au repository ?
14. Un utilisateur peut-il modifier la sélection de skills lors de la création d'une session, ou seulement les administrateurs ?
15. Le terme « admin » désigne-t-il uniquement `PLATFORM_ADMIN`, ou les `OWNER`/`ADMIN` d'organisation peuvent-ils aussi créer des agents limités à leur organisation ? La règle acquise est qu'un membre standard ne peut que sélectionner un agent publié et configurer les options repository autorisées.
16. Quelle politique s'applique lorsqu'un skill global est désactivé ou révoqué après la création d'une session ?
17. Pour une sortie suspecte ne correspondant pas à un secret connu, faut-il la bloquer systématiquement ou permettre sa consultation à un administrateur sécurité dans une quarantaine chiffrée ?

## P1 — Données, coûts et exploitation

18. Combien de runs simultanés, quelle durée maximale et quelle taille de workspace faut-il viser ?
19. Quelle rétention pour chats, événements filtrés, artefacts, audit et workspaces ?
20. Le workspace doit-il être un PVC permanent, une archive S3 restaurée à la demande, ou une combinaison des deux ?
21. Quels SLO, RPO et RTO sont attendus ?
22. Quelles destinations réseau l'agent doit-il joindre : Git, registries, package managers, endpoints de modèle, internet libre ?

## P2 — Après le premier vertical slice

23. Faut-il intégrer des déclencheurs GitHub/GitLab/Jira et commentaires de pull request ?
24. Faut-il un éditeur/diff interactif ou seulement chat et artefacts ?
25. Faut-il permettre plusieurs utilisateurs dans la même session en temps réel ?
26. Faut-il répartir les runs sur plusieurs clusters/régions ?
27. Faut-il publier/importer des skills depuis un catalogue externe ?

## Première question recommandée

Merci de commencer par préciser si les agents peuvent être créés uniquement par un **administrateur plateforme**, ou également par les `OWNER`/`ADMIN` d'une organisation pour leur propre périmètre. Ensuite, le lien ou l'image/version exacte d'OpenCode permettra de figer l'adaptateur, le format des événements et la stratégie de continuité.

