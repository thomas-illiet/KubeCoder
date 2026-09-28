# Chat et persistance des sessions

## Expérience attendue

Le chat est la vue principale d'une session. Il montre la conversation durable et les événements utiles de l'exécution sans déverser les logs bruts dans le fil. L'utilisateur peut quitter la page, revenir, retrouver le même historique et lancer un nouveau run dans le même contexte.

## Modèle de persistance

- `sessions` : organisation, repository, profil courant, titre, statut, timestamps et politique de rétention.
- `messages` : rôle, contenu structuré, ordre dans la session, run source et statut de consolidation.
- `runs` : état, snapshot, timestamps, raison de fin, métriques et identifiants Kubernetes.
- `run_events` : séquence, type, contenu déjà filtré, niveau de visibilité et clé d'idempotence ; aucune colonne de contenu brut par défaut.
- `artifacts` : nom logique, type, taille, digest, emplacement et politique d'accès.
- `runtime_checkpoints` : référence chiffrée/stockage objet vers l'état natif exporté par l'adaptateur.

Le contenu utilisateur et les réponses consolidées restent en PostgreSQL. Les gros flux et fichiers vont au stockage objet après le même pipeline de contrôle. Aucun contenu provenant du runtime ne devient durable ou visible avant redaction/quarantaine.

## Transport temps réel

- Commandes du navigateur par HTTP authentifié.
- Événements serveur par SSE, avec `id` monotone et reprise via `Last-Event-ID`.
- Heartbeat pour détecter les connexions mortes.
- Le serveur persiste un événement avant de le diffuser.
- Le serveur filtre un événement avant de le persister ; le frontend n'est jamais la frontière principale de masquage.
- À la reconnexion, l'API renvoie d'abord l'état canonique puis les événements manquants.

## Types d'événements visibles

- progression et état du run ;
- message partiel/final de l'assistant ;
- appel d'outil et résultat résumé ;
- demande de confirmation utilisateur ;
- fichier modifié ou artefact produit ;
- avertissement, erreur récupérable ou erreur terminale ;
- consommation indicative de ressources/tokens si disponible.

Les traces de diagnostic suivent le même filtrage avant stockage et ont une rétention courte. Une éventuelle quarantaine brute chiffrée est désactivée par défaut et nécessite une décision de sécurité explicite.

Lorsqu'un secret connu est détecté, le fil affiche `[SECRET_REDACTED]` sans donner son nom ou sa valeur. Lorsqu'une sortie est seulement suspecte, le fil affiche qu'un contenu a été retenu par la politique de sécurité ; seuls les administrateurs habilités voient les métadonnées de détection, jamais nécessairement le contenu brut.

## Reprise fonctionnelle

La reprise combine, dans cet ordre :

1. checkpoint natif du moteur si l'adaptateur le garantit ;
2. workspace persistant de la session ;
3. messages consolidés et snapshots précédents ;
4. nouveau message utilisateur.

Le niveau de fidélité est enregistré : `NATIVE`, `RECONSTRUCTED` ou `FRESH`. L'interface ne doit pas prétendre à une reprise native si le moteur ne la permet pas.

## Intégrité et rétention

- Le message utilisateur et le run sont créés atomiquement.
- Les événements sont append-only ; les projections de message peuvent être reconstruites.
- La suppression utilisateur devient d'abord une suppression logique, puis une purge asynchrone auditée.
- Les durées de conservation des chats, événements techniques filtrés, artefacts et workspaces sont configurables séparément.
- Un export de session contient messages, métadonnées, snapshots non secrets et index d'artefacts.

