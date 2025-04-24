# FileWeightLoss CLI
*N'attendez plus pour faire perdre du poids à vos fichiers.*

> Retrouvez la version GUI [ici](https://github.com/el2zay/fileweightloss)<br>

## À quoi sert FileWeightLoss
Fileweightloss permet de compresser un fichier vidéo ou audio dans certains formats courants (.mp4, .mov, .mp3, ...) sans réduire la résolution, et sans faire de trop grosses modifications dans la qualité visuelle ou sonore.

Il utilise [ffmpeg](https://ffmpeg.org/download.html), ce dernier doit obligatoirement être installé sur votre ordinateur.

## TODO

- [X] Meilleure gestion de certaines erreurs
- [X] Argument pour afficher les logs ffmpeg au lieu d'une animation de chargement
- [X] Argument pour n'utiliser qu'une seule tentative de compression
- [X] Pouvoir afficher une sortie différente en JSON pour faciliter l'automatisation
- [ ] Notification a la fin de la compression si elle tourne depuis plus d'une minute [prêt mais bug sur windows 11](https://github.com/gen2brain/beeep/issues/57)
- [ ] Compresser plusieurs fichiers en une commande
- [ ] Prompt pour installer facilement ffmpeg si il n'est pas sur la machine.
- [ ] Argument pour choisir le niveau de compression
- [ ] Installateur graphique pour Windows
- [ ] Publication dans le MS Store et gestionnaire de paquets

