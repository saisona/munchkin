# 📜 Règles du jeu Munchkin — Spécification de développement

> **Version** : 2.0
> **Usage** : Document de référence pour le développement assisté par IA
> **Scope** : Jeu de base uniquement (pas d'extensions)
> **Sources** : Règles officielles Steve Jackson Games, [regledujeu.fr](https://www.regledujeu.fr/munchkin/), [wiki Munchkin](https://fr.wikipedia.org/wiki/Munchkin_(jeu))

---

## Table des matières

1. [Glossaire](#1--glossaire)
2. [Objectif & condition de victoire](#2--objectif--condition-de-victoire)
3. [Joueurs & Matériel](#3--joueurs--matériel)
4. [Types de cartes — Taxonomie complète](#4--types-de-cartes--taxonomie-complète)
5. [Personnage — Structure de données](#5--personnage--structure-de-données)
6. [Mise en place](#6--mise-en-place)
7. [Déroulement d'un tour — Machine à états](#7--déroulement-dun-tour--machine-à-états)
8. [Système de combat — Algorithme détaillé](#8--système-de-combat--algorithme-détaillé)
9. [Équipements — Règles de slot et restrictions](#9--équipements--règles-de-slot-et-restrictions)
10. [Économie — Vente d'objets](#10--économie--vente-dobjets)
11. [Mort du personnage](#11--mort-du-personnage)
12. [Interactions entre joueurs hors combat](#12--interactions-entre-joueurs-hors-combat)
13. [Règles de priorité & résolution de conflits](#13--règles-de-priorité--résolution-de-conflits)
14. [Contraintes & cas limites](#14--contraintes--cas-limites)
15. [Modèle de données — Schéma formel](#15--modèle-de-données--schéma-formel)
16. [Machine à états globale — Diagramme](#16--machine-à-états-globale--diagramme)

---

## 1 — Glossaire

> **Objectif** : Éviter toute ambiguïté dans les échanges avec Claude. Chaque terme a **une seule définition**.

| Terme | Définition exacte |
|-------|-------------------|
| **Joueur actif** | Le joueur dont c'est le tour en cours |
| **Joueur inactif** | Tout joueur qui n'est pas le joueur actif |
| **Niveau** | Entier de 1 à 10 représentant la puissance de base du personnage |
| **Force de combat** | Valeur calculée = `niveau + bonus_équipements + bonus_temporaires` |
| **Bonus temporaire** | Tout bonus appliqué uniquement pour la durée d'un combat (cartes Action jouées depuis la main) |
| **Équipement porté** | Carte Objet posée face visible devant le joueur, occupant un slot, et dont le bonus est actif |
| **Équipement transporté** | Carte Objet posée face visible devant le joueur mais **non équipée** (ne donne pas de bonus) car un slot est déjà occupé ou une restriction empêche l'utilisation |
| **Main** | Ensemble des cartes qu'un joueur tient en main (non visibles par les autres). Maximum **5 cartes en fin de tour** |
| **Pioche Donjon** | Paquet de cartes contenant : Monstres, Malédictions, Races, Classes, Actions Donjon |
| **Pioche Trésor** | Paquet de cartes contenant : Objets (équipements), Actions Trésor |
| **Défausse Donjon** | Pile face visible des cartes Donjon jouées/défaussées |
| **Défausse Trésor** | Pile face visible des cartes Trésor jouées/défaussées |
| **Gros objet** | Objet marqué "Gros". Un joueur ne peut porter qu'**un seul** Gros objet (sauf Nain) |
| **Monstre errant** | Carte Monstre jouée depuis la main d'un joueur inactif pour **s'ajouter** au combat en cours |
| **Chercher des ennuis** | Action du joueur actif consistant à jouer un Monstre depuis sa main pour le combattre volontairement |
| **Piller la salle** | Action de piocher une carte Donjon face cachée quand aucun combat n'a eu lieu |
| **Charité** | Phase de fin de tour : le joueur doit réduire sa main à 5 cartes maximum |
| **Fuite** | Tentative d'échapper à un monstre en lançant un d6 (≥ 5 = succès) |
| **Mort (game over)** | État temporaire quand la fuite échoue. Le joueur perd ses cartes en jeu mais conserve son niveau |
| **Allié** | Joueur inactif qui accepte d'aider le joueur actif dans un combat, après négociation |

---

## 2 — Objectif & condition de victoire

```
SI joueur.niveau == 10
  ET ce niveau a été atteint par la résolution victorieuse d'un combat contre un monstre
ALORS ce joueur gagne la partie immédiatement.
```

### Règles strictes

- Le niveau 10 **NE PEUT PAS** être atteint par :
  - La vente d'objets
  - Une carte Action donnant des niveaux
  - Toute autre source que le combat
- Si un joueur est niveau 9 et joue une carte "Gagne 2 niveaux" → il passe à **10** ? **NON** — il reste à **9**. La carte est perdue/défaussée.
- Si un joueur est niveau 9 et tue un monstre qui donne 2 niveaux → il passe à **10** → **VICTOIRE** ✅
- En cas d'égalité théorique (deux joueurs atteignent 10 en même temps via aide) : seul le **joueur actif** gagne

---

## 3 — Joueurs & Matériel

| Paramètre | Valeur | Note technique |
|-----------|--------|----------------|
| Joueurs min | 3 | En dessous, les règles de négociation perdent leur intérêt |
| Joueurs max | 6 | Au-delà, les tours sont trop longs |
| Cartes Donjon | 95 | Jeu de base |
| Cartes Trésor | 73 | Jeu de base |
| Dé | 1d6 | Utilisé uniquement pour la fuite et certains effets de cartes |

---

## 4 — Types de cartes — Taxonomie complète

### 4.1 — Cartes Donjon

```
CarteDonjon
├── Monstre
│   ├── nom: string
│   ├── niveau: int
│   ├── bonus_contre: Race | Classe | null  // ex: +3 contre Elfes
│   ├── malus_fuite: int                     // modificateur de fuite (peut être 0)
│   ├── punition_fuite: Effet                // ce qui se passe si fuite échouée
│   ├── nb_tresors: int                      // nb de cartes Trésor en récompense
│   └── niveaux_gagnes: int                  // généralement 1, parfois 2+
│
├── Malédiction
│   ├── nom: string
│   └── effet: Effet                         // ex: perd ton casque, perd 1 niveau
│
├── Race
│   ├── nom: enum(ELFE, NAIN, HALFELIN)
│   └── capacite: Capacité[]
│
├── Classe
│   ├── nom: enum(GUERRIER, VOLEUR, MAGE, PRETRE)
│   └── capacite: Capacité[]
│
├── SangMele                                 // permet 2 races
├── SuperMunchkin                            // permet 2 classes
│
└── ActionDonjon
    ├── nom: string
    ├── quand_jouable: MomentJeu             // voir section 13
    └── effet: Effet
```

### 4.2 — Cartes Trésor

```
CarteTresor
├── Objet
│   ├── nom: string
│   ├── bonus: int                          // bonus de combat quand porté
│   ├── valeur_or: int                      // prix de vente (par multiple de 100)
│   ├── taille: enum(NORMAL, GROS)
│   ├── slot: enum(TETE, ARMURE, PIEDS, MAIN_1, MAIN_2, DEUX_MAINS, AUCUN)
│   ├── restriction_race: Race | null
│   ├── restriction_classe: Classe | null
│   └── restriction_sexe: Sexe | null
│
└── ActionTresor
    ├── nom: string
    ├── quand_jouable: MomentJeu
    └── effet: Effet
```

### 4.3 — Énumération `MomentJeu`

> Critique pour le développement : définit **quand** chaque carte peut être jouée.

| Valeur | Signification |
|--------|---------------|
| `PENDANT_SON_TOUR` | Jouable uniquement par le joueur actif, hors combat |
| `PENDANT_UN_COMBAT` | Jouable par n'importe quel joueur pendant un combat |
| `A_TOUT_MOMENT` | Jouable à n'importe quel moment du jeu par n'importe qui |
| `EN_REPONSE` | Jouable en réaction à une autre carte ou un événement |

### 4.4 — Règles de jeu des cartes depuis la main

| Type de carte | Quand peut-on la jouer depuis la main ? |
|---------------|----------------------------------------|
| Race | `PENDANT_SON_TOUR` (hors combat) — remplace la race actuelle |
| Classe | `PENDANT_SON_TOUR` (hors combat) — remplace la classe actuelle |
| Sang-Mêlé / Super Munchkin | `PENDANT_SON_TOUR` (hors combat) |
| Monstre (Chercher des ennuis) | Phase 2 du tour du joueur actif uniquement |
| Monstre (Monstre errant) | `PENDANT_UN_COMBAT` — ajoute un monstre au combat |
| Malédiction (depuis la main) | `A_TOUT_MOMENT` — le joueur choisit la cible |
| Objet (équiper) | `PENDANT_SON_TOUR` (hors combat) |
| Action Trésor (bonus combat) | `PENDANT_UN_COMBAT` |
| Action Trésor (autre) | Selon `quand_jouable` de la carte |

---

## 5 — Personnage — Structure de données

```
Personnage {
  niveau: int                     // 1 à 10, initialisé à 1
  race: Race | null               // null = Humain (défaut)
  race2: Race | null              // non-null seulement si SangMele actif
  classe: Classe | null           // null = aucune classe
  classe2: Classe | null          // non-null seulement si SuperMunchkin actif
  sexe: enum(MASCULIN, FEMININ)   // défini au début, peut changer par malédiction
  sang_mele: bool                 // a la carte Sang-Mêlé en jeu
  super_munchkin: bool            // a la carte Super Munchkin en jeu
  equipements_portes: Objet[]     // slots occupés, bonus actifs
  equipements_transportes: Objet[] // en jeu mais pas de bonus
  main: Carte[]                   // cartes en main (max 5 fin de tour)
  est_mort: bool                  // état temporaire
}
```

### 5.1 — Capacités des Races (jeu de base)

| Race | Capacité |
|------|----------|
| **Humain** (défaut, pas de carte) | Aucune capacité spéciale. Peut se défausser d'une race pour redevenir Humain |
| **Elfe** | Gagne +1 niveau quand il **aide** un autre joueur à tuer un monstre |
| **Nain** | Peut porter **plusieurs objets Gros** (ignore la limite de 1 Gros objet) |
| **Halfelin** | Peut **relancer le dé de fuite** une fois par combat. Bonus de +1 à la fuite (seuil effectif ≥ 4) |

### 5.2 — Capacités des Classes (jeu de base)

| Classe | Capacité |
|--------|----------|
| **Aucune** (défaut) | Aucune capacité spéciale |
| **Guerrier** | Peut défausser jusqu'à **3 cartes de sa main** pendant un combat pour gagner **+1 par carte** défaussée |
| **Voleur** | Peut tenter de **voler un objet** porté par un autre joueur (jet de d6, ≥ 4 = succès). Échec = perd 1 niveau. **1 tentative par tour** |
| **Mage** | Peut **défausser jusqu'à 3 cartes de sa main** pendant un combat pour un effet de charme (modifier le monstre ou fuir automatiquement selon la carte — à préciser par carte) |
| **Prêtre** | Peut **défausser sa main entière** pour ressusciter un joueur mort avant la phase de pillage du cadavre |

---

## 6 — Mise en place

```
PROCÉDURE mise_en_place(joueurs[]):
  1. Mélanger la pioche Donjon
  2. Mélanger la pioche Trésor
  3. POUR CHAQUE joueur DANS joueurs:
     a. joueur.niveau = 1
     b. joueur.race = null          // Humain
     c. joueur.classe = null        // Aucune
     d. joueur.sexe = sexe du joueur réel (ou choix libre en jeu vidéo)
     e. Piocher 4 cartes de la pioche Donjon → joueur.main
     f. Piocher 4 cartes de la pioche Trésor → joueur.main
     // main initiale = 8 cartes (exception à la limite de 5,
     //                            la charité ne s'applique pas au premier tour)
  4. POUR CHAQUE joueur (avant le premier tour):
     - Le joueur PEUT poser immédiatement ses cartes Race, Classe et Objets
       (s'il remplit les conditions) depuis sa main
  5. Déterminer le premier joueur (aléatoire ou plus récent anniversaire)
  6. Sens du jeu : sens horaire
```

> ⚠️ La limite de 5 cartes ne s'applique qu'**en fin de tour**, pas durant la mise en place.

---

## 7 — Déroulement d'un tour — Machine à états

```
                      ┌─────────────────┐
                      │  DÉBUT DU TOUR  │
                      │  (joueur actif)  │
                      └────────┬────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │  PHASE 1 : OUVRIR   │
                    │    LA PORTE          │
                    │ (piocher 1 Donjon   │
                    │  face visible)       │
                    └────────┬────────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
         [MONSTRE]     [MALÉDICTION]    [AUTRE]
              │              │              │
              │         S'applique     Prendre en
              │         immédiatement   main OU
              │         au joueur       jouer
              │         actif           immédiatement
              │              │              │
              ▼              │              │
     ┌────────────────┐     │              │
     │    COMBAT      │     │              │
     │  (voir §8)     │     │              │
     └───────┬────────┘     │              │
             │              ▼              ▼
             │    ┌───────────────────────────┐
             │    │  PHASE 2 : CHERCHER DES   │
             │    │  ENNUIS (optionnel)        │
             │    │  Jouer un Monstre depuis   │
             │    │  la main → COMBAT          │
             │    └────────────┬──────────────┘
             │                 │
             │         ┌──────┴──────┐
             │         │             │
             │     [COMBAT]   [PAS DE COMBAT]
             │         │             │
             │         │             ▼
             │         │   ┌──────────────────┐
             │         │   │  PHASE 3 :       │
             │         │   │  PILLER LA SALLE │
             │         │   │  (piocher 1      │
             │         │   │  Donjon FACE     │
             │         │   │  CACHÉE en main) │
             │         │   └────────┬─────────┘
             │         │            │
             ▼         ▼            ▼
        ┌─────────────────────────────────┐
        │  PHASE 4 : CHARITÉ              │
        │  SI main.taille > 5 :           │
        │    donner surplus au joueur     │
        │    du niveau le plus bas        │
        │    (si c'est soi-même :         │
        │     défausser)                  │
        └──────────────┬──────────────────┘
                       │
                       ▼
              ┌────────────────┐
              │  FIN DU TOUR   │
              │  → joueur      │
              │    suivant     │
              └────────────────┘
```

### 7.1 — Règles précises par phase

#### Phase 1 — Ouvrir la Porte

1. Le joueur actif pioche **1 carte Donjon** et la retourne **face visible** (tous les joueurs la voient)
2. Résolution immédiate selon le type :
   - **Monstre** → Combat obligatoire (sauf cartes spéciales)
   - **Malédiction** → Effet immédiat sur le joueur actif. La carte est défaussée après effet
   - **Race / Classe / Action** → Le joueur **choisit** : la garder en main **OU** la jouer immédiatement

#### Phase 2 — Chercher des Ennuis (OPTIONNEL)

- **Condition** : aucun combat n'a eu lieu en phase 1
- Le joueur actif **peut** jouer **1 carte Monstre depuis sa main** pour déclencher un combat
- S'il ne le fait pas → passer à la phase 3

#### Phase 3 — Piller la Salle (OPTIONNEL)

- **Condition** : aucun combat n'a eu lieu (ni en phase 1 ni en phase 2)
- Le joueur actif pioche **1 carte Donjon face cachée** → elle va directement en main (personne ne la voit)

#### Phase 4 — Charité

- **Condition** : la main du joueur actif contient **plus de 5 cartes**
- Il **doit** réduire sa main à **5 cartes maximum**
- Les cartes excédentaires sont **données au joueur du niveau le plus bas**
- Si **plusieurs joueurs sont à égalité** au niveau le plus bas → le joueur actif **choisit** la répartition entre eux
- Si le joueur actif **est lui-même** le plus bas niveau (ou à égalité) → il **défausse** les cartes excédentaires dans les défausses appropriées (Donjon ou Trésor selon le type)

---

## 8 — Système de combat — Algorithme détaillé

### 8.1 — Déclenchement

Un combat commence quand :
- Un Monstre est révélé en phase 1 (Ouvrir la Porte), **OU**
- Le joueur actif joue un Monstre depuis sa main (Chercher des Ennuis, phase 2)

### 8.2 — Fenêtre d'interaction

Après le déclenchement du combat, **avant la résolution** :

```
BOUCLE interaction_combat:
  TANT QUE au moins un joueur veut agir:
    N'IMPORTE QUEL joueur peut:
      - Jouer une carte bonus de combat (pour ou contre)
      - Jouer un Monstre Errant (ajoute un monstre au combat)
      - Jouer une Malédiction depuis la main
      - Proposer / accepter une alliance (joueur actif + 1 allié max)
      - Utiliser une capacité de classe (ex: Guerrier défausse des cartes)
    
    Le joueur actif peut:
      - Accepter ou refuser une offre d'alliance
      - Négocier le partage des trésors avec un allié potentiel
      - Utiliser des objets à usage unique
```

> ⚠️ **Un seul allié maximum** peut aider le joueur actif par combat.

### 8.3 — Calcul de la force

```
force_joueur_actif =
    joueur_actif.niveau
  + somme(objet.bonus POUR objet DANS joueur_actif.equipements_portes)
  + somme(bonus_temporaires appliqués au joueur actif)
  + bonus_classe_si_applicable

force_allie =                          // 0 si pas d'allié
    allie.niveau
  + somme(objet.bonus POUR objet DANS allie.equipements_portes)
  + somme(bonus_temporaires appliqués à l'allié)
  + bonus_classe_si_applicable

force_totale_joueurs = force_joueur_actif + force_allie

force_monstre =
    monstre.niveau
  + somme(modificateurs joués par les joueurs inactifs)
  + monstre.bonus_contre SI joueur_actif.race correspond
  + somme(niveaux des monstres errants ajoutés)
```

### 8.4 — Résolution

```
SI force_totale_joueurs > force_monstre:
    → VICTOIRE
SINON:   // force_totale_joueurs ≤ force_monstre
    → DÉFAITE (tentative de fuite obligatoire)
```

> ⚠️ **Strictement supérieur**. L'égalité = défaite.

### 8.5 — Victoire

```
PROCÉDURE victoire(joueur_actif, allie?, monstre):
  1. joueur_actif.niveau += monstre.niveaux_gagnes
  
  2. SI allie != null ET allie.race == ELFE:
       allie.niveau += 1
  
  3. Piocher monstre.nb_tresors cartes de la pioche Trésor
  
  4. SI allie != null:
       Répartir les trésors selon l'accord négocié
       (le joueur actif pioche en premier et choisit en premier
        sauf accord contraire)
     SINON:
       Tous les trésors vont au joueur actif
  
  5. Défausser la carte Monstre
  
  6. VÉRIFIER condition de victoire (niveau 10)
```

### 8.6 — Défaite & Fuite

```
PROCÉDURE fuite(joueur, monstre):
  // Chaque joueur impliqué (actif + allié) fuit séparément
  
  POUR CHAQUE joueur_implique DANS [joueur_actif, allie?]:
    resultat_de = lancer_d6()
    resultat_final = resultat_de + bonus_fuite - monstre.malus_fuite
    
    SI joueur_implique.race == HALFELIN:
      SI resultat_final < 5:
        resultat_de = lancer_d6()  // relance
        resultat_final = resultat_de + bonus_fuite - monstre.malus_fuite
    
    SI resultat_final >= 5:
      → Fuite réussie. Aucune conséquence pour ce joueur.
    SINON:
      → Fuite échouée. Appliquer monstre.punition_fuite à ce joueur.
```

### 8.7 — Punitions de fuite échouée (exemples)

| Punition | Effet |
|----------|-------|
| Mort | Voir section 11 |
| Perte d'objet | Perdre l'objet de plus grande valeur (ou au choix si égalité) |
| Perte de niveau(x) | Perdre N niveaux (minimum 1) |
| Perte de cartes en main | Défausser N cartes au hasard |
| Perte de classe/race | Défausser la carte Race ou Classe |

### 8.8 — Combat contre plusieurs monstres

Quand un Monstre Errant est joué :

- Les forces de **tous les monstres s'additionnent** pour le calcul
- En cas de victoire : le joueur gagne **+1 niveau par monstre** tué et **cumule les trésors**
- En cas de fuite : le joueur doit fuir **chaque monstre séparément** (un jet de dé par monstre)
- Chaque fuite échouée applique la punition du monstre correspondant

---

## 9 — Équipements — Règles de slot et restrictions

### 9.1 — Slots disponibles

| Slot | Quantité max | Exemples |
|------|-------------|----------|
| `TETE` | 1 | Casque de Corna, Couvre-chef Enflammé |
| `ARMURE` | 1 | Armure Flamboyante, Mithril |
| `PIEDS` | 1 | Bottes de Course, Sandales de Protection |
| `MAIN` | 2 mains au total | Épée, Bouclier (1 main chacun) |
| `DEUX_MAINS` | Occupe les 2 mains | Arc, Bâton (incompatible avec tout autre objet de main) |
| `AUCUN` | Pas de limite de slot | Amulettes, anneaux, etc. |

### 9.2 — Restriction : Gros objet

- Un joueur ne peut **porter** (avoir équipé) qu'**un seul** objet Gros à la fois
- **Exception** : le **Nain** peut porter **plusieurs** objets Gros
- Un joueur peut **transporter** (avoir en jeu mais pas équipé) un objet Gros supplémentaire sans bénéficier de son bonus

### 9.3 — Restrictions conditionnelles

```
FONCTION peut_equiper(joueur, objet) → bool:
  SI objet.restriction_race != null ET joueur.race != objet.restriction_race:
    RETOURNER false
  SI objet.restriction_classe != null ET joueur.classe != objet.restriction_classe:
    RETOURNER false
  SI objet.restriction_sexe != null ET joueur.sexe != objet.restriction_sexe:
    RETOURNER false
  SI slot_deja_occupe(joueur, objet.slot):
    RETOURNER false
  SI objet.taille == GROS ET joueur a déjà un gros objet ET joueur.race != NAIN:
    RETOURNER false
  RETOURNER true
```

### 9.4 — Porter vs Transporter

| État | Bonus actif ? | Visible ? | Peut être vendu ? | Peut être volé ? |
|------|:---:|:---:|:---:|:---:|
| **Porté** (équipé) | ✅ Oui | ✅ Oui | ✅ Oui | ✅ Oui |
| **Transporté** (en jeu, pas équipé) | ❌ Non | ✅ Oui | ✅ Oui | ✅ Oui |
| **En main** | ❌ Non | ❌ Non | ❌ Non | ❌ Non |

### 9.5 — Changement d'équipement

- Un joueur peut **changer ses équipements** uniquement **pendant son tour, hors combat**
- Changer d'équipement = retirer un objet porté (il devient transporté) et en équiper un autre
- Un objet retiré peut être **vendu, donné, ou gardé en transport**

---

## 10 — Économie — Vente d'objets

```
PROCÉDURE vendre_objets(joueur, objets_a_vendre[]):
  // Uniquement pendant son tour, hors combat
  
  total_or = somme(objet.valeur_or POUR objet DANS objets_a_vendre)
  niveaux_gagnes = total_or / 1000  // division entière, arrondi vers le bas
  reste = total_or % 1000            // PAS de monnaie rendue, le reste est perdu
  
  SI joueur.niveau + niveaux_gagnes >= 10:
    niveaux_gagnes = 9 - joueur.niveau  // PLAFONNER à 9, jamais 10 par la vente
  
  joueur.niveau += niveaux_gagnes
  
  POUR CHAQUE objet DANS objets_a_vendre:
    deplacer objet → défausse Trésor
```

> ⚠️ On peut vendre **plusieurs objets en même temps** pour atteindre le seuil de 1000. Les objets en main **NE PEUVENT PAS** être vendus (seulement les objets en jeu : portés ou transportés).

---

## 11 — Mort du personnage

```
PROCÉDURE mort(joueur):
  1. joueur.est_mort = true
  
  2. Le joueur CONSERVE:
     - Son niveau
     - Sa race (carte en jeu)
     - Sa classe (carte en jeu)
  
  3. Le joueur PERD:
     - TOUS ses équipements (portés et transportés)
     - TOUTES ses cartes en main
     → Ces cartes sont étalées face visible
  
  4. PILLAGE DU CADAVRE:
     Les autres joueurs, en partant du joueur avec le plus haut niveau
     puis en sens horaire, prennent chacun 1 carte jusqu'à ce que
     toutes les cartes soient prises.
     S'il reste des cartes que personne ne veut → défausser.
  
  5. RÉSURRECTION:
     Au prochain tour du joueur mort:
     - Piocher 4 cartes Donjon
     - Piocher 4 cartes Trésor
     - Le joueur peut immédiatement équiper ce qu'il peut
     - Son tour commence normalement
  
  6. joueur.est_mort = false
```

---

## 12 — Interactions entre joueurs hors combat

### 12.1 — Échange / Don d'objets

- Un joueur **peut donner** des objets en jeu (portés ou transportés) à un autre joueur **pendant son propre tour, hors combat**
- Le receveur **doit respecter** les règles de slot et restrictions pour équiper l'objet — sinon il le transporte
- Aucun **échange direct** carte-contre-carte n'est autorisé dans les règles de base

### 12.2 — Malédictions depuis la main

- Un joueur peut jouer une carte Malédiction **depuis sa main** sur **n'importe quel joueur** à **n'importe quel moment**
- La malédiction s'applique immédiatement
- Différence avec la malédiction piochée en phase 1 : ici, le joueur **choisit sa cible**

### 12.3 — Vol (Voleur uniquement)

```
PROCÉDURE tenter_vol(voleur, cible, objet_cible):
  // 1 tentative par tour du voleur uniquement
  // L'objet doit être en jeu (porté ou transporté) chez la cible
  
  resultat = lancer_d6()
  
  SI resultat >= 4:
    → Vol réussi. L'objet passe au Voleur (porté si possible, sinon transporté)
  SINON:
    → Vol échoué. Le Voleur perd 1 niveau (minimum 1).
```

---

## 13 — Règles de priorité & résolution de conflits

### 13.1 — Ordre de résolution

Quand plusieurs joueurs veulent agir en même temps :

1. Le **joueur actif** a toujours la priorité pour agir en premier
2. Ensuite, les joueurs agissent **dans le sens horaire** à partir du joueur actif
3. Une carte jouée **en réponse** à une autre se résout **avant** la carte initiale (pile LIFO)

### 13.2 — Pioche vide

```
SI pioche_donjon est vide:
  Mélanger la défausse Donjon → nouvelle pioche Donjon

SI pioche_tresor est vide:
  Mélanger la défausse Trésor → nouvelle pioche Trésor
```

### 13.3 — Niveau minimum

- Le niveau d'un joueur **ne peut jamais descendre en dessous de 1**
- Si un effet ferait descendre sous 1 → le joueur reste au niveau 1

### 13.4 — Timing des cartes Race/Classe

- Jouer une nouvelle Race **remplace** la race actuelle (l'ancienne carte est défaussée)
- Jouer une nouvelle Classe **remplace** la classe actuelle
- Se **défausser** volontairement d'une Race → redevient Humain
- Se **défausser** volontairement d'une Classe → n'a plus de classe

---

## 14 — Contraintes & cas limites

| Cas limite | Règle |
|-----------|-------|
| Joueur actif est niveau 9 et reçoit un bonus de niveau hors combat | Reste à 9 (ne peut pas atteindre 10 hors combat monstre) |
| Deux monstres en combat, le joueur ne peut fuir que l'un | Il subit la punition de celui qu'il n'a pas pu fuir |
| Joueur mort pendant un autre tour que le sien | Il attend son prochain tour pour ressusciter |
| Allié meurt en combat (fuite ratée) | L'allié subit la punition. Le joueur actif fuit séparément |
| Joueur n'a pas de carte Monstre pour Chercher des Ennuis | Il passe directement à Piller la Salle |
| Un joueur joue une Race/Classe qui invalide un objet équipé | L'objet devient immédiatement "transporté" (perd son bonus) |
| Charité avec plusieurs joueurs au même niveau le plus bas | Le joueur actif répartit les cartes entre eux comme il veut |
| Vente d'objets : total < 1000 or | Pas de niveau gagné, les objets sont quand même défaussés |
| Monstre de niveau 1, joueur de niveau 1 sans bonus | Égalité → défaite (strict supérieur requis) |

---

## 15 — Modèle de données — Schéma formel

```
// ==================== ENUMS ====================

enum PiocheType { DONJON, TRESOR }
enum Slot { TETE, ARMURE, PIEDS, MAIN, DEUX_MAINS, AUCUN }
enum Taille { NORMAL, GROS }
enum Sexe { MASCULIN, FEMININ }
enum RaceType { HUMAIN, ELFE, NAIN, HALFELIN }
enum ClasseType { AUCUNE, GUERRIER, VOLEUR, MAGE, PRETRE }
enum PhaseType { OUVRIR_PORTE, CHERCHER_ENNUIS, PILLER_SALLE, CHARITE }
enum CombatResultat { VICTOIRE, FUITE_REUSSIE, FUITE_ECHOUEE }
enum MomentJeu { PENDANT_SON_TOUR, PENDANT_UN_COMBAT, A_TOUT_MOMENT, EN_REPONSE }

// ==================== CARTES ====================

interface Carte {
  id: unique
  nom: string
  type_pioche: PiocheType
  description: string
}

interface CarteMonstre extends Carte {
  niveau: int
  bonus_contre_race: RaceType | null
  bonus_contre_classe: ClasseType | null
  valeur_bonus_contre: int              // ex: +3
  malus_fuite: int                      // 0 = pas de malus
  punition: Punition
  nb_tresors: int
  niveaux_gagnes: int                   // généralement 1
}

interface CarteObjet extends Carte {
  bonus_combat: int
  valeur_or: int                        // multiple de 100
  taille: Taille
  slot: Slot
  nb_mains: int                         // 0, 1 ou 2 (si slot == MAIN ou DEUX_MAINS)
  restriction_race: RaceType | null
  restriction_classe: ClasseType | null
  restriction_sexe: Sexe | null
}

interface CarteMalediction extends Carte {
  effet: Effet
}

interface CarteRace extends Carte {
  race: RaceType
}

interface CarteClasse extends Carte {
  classe: ClasseType
}

interface CarteAction extends Carte {
  moment_jouable: MomentJeu
  effet: Effet
}

// ==================== JOUEUR ====================

interface Joueur {
  id: unique
  nom: string
  niveau: int                           // [1, 10]
  sexe: Sexe
  race_principale: CarteRace | null     // null = Humain
  race_secondaire: CarteRace | null     // non-null seulement si sang_mele
  classe_principale: CarteClasse | null // null = aucune
  classe_secondaire: CarteClasse | null // non-null seulement si super_munchkin
  a_sang_mele: bool
  a_super_munchkin: bool
  equipements_portes: CarteObjet[]
  equipements_transportes: CarteObjet[]
  main: Carte[]
  est_mort: bool
}

// ==================== ÉTAT DU JEU ====================

interface EtatJeu {
  joueurs: Joueur[]                     // 3 à 6
  joueur_actif_index: int
  pioche_donjon: Carte[]                // pile face cachée
  pioche_tresor: Carte[]                // pile face cachée
  defausse_donjon: Carte[]              // pile face visible
  defausse_tresor: Carte[]              // pile face visible
  phase_courante: PhaseType
  combat_en_cours: Combat | null
  sens_jeu: HORAIRE                     // toujours horaire
}

interface Combat {
  monstre_principal: CarteMonstre
  monstres_errants: CarteMonstre[]      // monstres ajoutés
  modificateurs_monstre: int            // bonus ajoutés par les joueurs inactifs
  joueur_actif: Joueur
  allie: Joueur | null
  bonus_temporaires_joueur: int
  bonus_temporaires_allie: int
  accord_tresor: AccordTresor | null    // négociation en cours/conclue
  phase_combat: enum(INTERACTION, RESOLUTION, FUITE, TERMINE)
}
```

---

## 16 — Machine à états globale — Diagramme

```
[INITIALISATION]
      │
      ▼
[DISTRIBUTION_CARTES] ──► Chaque joueur reçoit 4 Donjon + 4 Trésor
      │
      ▼
[EQUIPER_INITIAL] ──► Les joueurs posent Race/Classe/Objets s'ils le souhaitent
      │
      ▼
[DEBUT_TOUR] ◄─────────────────────────────────────────┐
      │                                                  │
      ▼                                                  │
[PHASE_OUVRIR_PORTE]                                     │
      │                                                  │
      ├── Monstre → [COMBAT]                             │
      │                  │                               │
      │           ┌──────┴──────┐                        │
      │      [VICTOIRE]    [DEFAITE]                     │
      │           │             │                        │
      │           │        [FUITE]                       │
      │           │          │     │                     │
      │           │     [REUSSIE] [ECHOUEE]              │
      │           │                  │                   │
      │           │             [PUNITION]               │
      │           │                  │                   │
      │           │    ┌─────────────┘                   │
      │           │    │                                 │
      │           ▼    ▼                                 │
      │      [VERIF_VICTOIRE_PARTIE]                     │
      │           │                                      │
      │       [Niveau 10?]──Oui──► [FIN_PARTIE]         │
      │           │                                      │
      │          Non                                     │
      │           │                                      │
      ├── Malédiction → appliquer effet                  │
      │           │                                      │
      ├── Autre → en main ou jouer                       │
      │           │                                      │
      ▼           ▼                                      │
[PHASE_CHERCHER_ENNUIS] ─── Monstre joué? ──► [COMBAT]  │
      │                                                  │
      │ (pas de combat)                                  │
      ▼                                                  │
[PHASE_PILLER_SALLE] ──► piocher 1 Donjon face cachée   │
      │                                                  │
      ▼                                                  │
[PHASE_CHARITE] ──► réduire main à 5 max                 │
      │                                                  │
      ▼                                                  │
[FIN_TOUR] ──► joueur_actif_index++ (modulo nb joueurs) ─┘
```

---

> **Ce document fait autorité** pour toutes les décisions de développement. En cas de doute sur une règle non couverte ici, demander une clarification plutôt que d'improviser.
