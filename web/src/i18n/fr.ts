import type { en } from './en'

export const fr: typeof en = {
  app: { title: 'Open Niscat', edition: 'Données : {edition}', loading: 'Chargement…', retry: 'Réessayer', notFound: 'Page introuvable' },
  error: { not_found: 'Introuvable.', invalid: 'Requête invalide.', internal: 'Erreur du serveur.', network: 'Le serveur est injoignable.' },
  nav: { home: 'Accueil', catalogs: 'Catalogues', cart: 'Panier', vehicle: 'Véhicule' },
  home: {
    vinLabel: 'VIN', vinHint: 'VIN complet, ou ses 6 derniers caractères (ou plus)', find: 'Identifier',
    browse: 'Parcourir les catalogues', history: 'Consultés récemment', searchLabel: 'Rechercher des pièces ou des sections',
  },
  scope: {
    none: 'Aucun véhicule', choose: 'Choisir un véhicule', use: 'Utiliser ce véhicule', clear: 'Retirer le véhicule',
    saved: 'Véhicule mémorisé', sheet: 'Fiche véhicule', fromLink: 'Véhicule du lien',
  },
  vin: { candidates: 'VIN correspondants', notFound: 'Aucun véhicule pour ce VIN.' },
  vehicle: {
    vin: 'VIN', model: 'Code modèle', prodDate: 'Production', vinCount: 'VIN avec ce code modèle',
    documents: 'Documents', groups: 'Index général', attributes: 'Caractéristiques', catalog: 'Catalogue',
  },
  catalogs: { title: 'Catalogues', period: 'Période', drive: 'Conduite' },
  models: { title: 'Codes modèle', filter: 'Filtrer', use: 'Utiliser', vins: 'VIN', count: '{n} codes modèle' },
  group: {
    sections: 'Sections', showAll: 'Afficher les sections non applicables', notApplicable: 'Non applicable',
    filtered: 'Sections pour {caption}', clearFilter: 'Toutes les sections',
  },
  section: {
    notApplicable: 'Cette section ne s’applique pas au véhicule actif.', prev: 'Précédente', next: 'Suivante',
    print: 'Imprimer', drawing: 'Vue éclatée', info: 'Infos', select: 'Sélectionnez un repère sur le dessin ou dans la liste.',
    fit: 'Ajuster', zoomIn: 'Zoomer', zoomOut: 'Dézoomer', group: 'Groupe',
  },
  part: {
    mark: 'Marque', item: 'Repère', reference: 'Numéro de pièce', description: 'Description', qty: 'Qté', period: 'Période',
    alternative: 'Numéro de pièce alternative', latest: 'Dernière référence connue', ica: 'ICA', app: 'Applicable à modèle',
    spec: 'Spécification', pnc: 'Obs. (PNC)', kd: 'K.D.', outOfPeriod: 'Hors de la date de production du véhicule',
    occurrences: 'Utilisée dans', replaces: 'Remplace', replacedBy: 'Remplacée par', series: 'Catalogue', section: 'Section',
  },
  search: {
    placeholder: 'Rechercher…', parts: 'Pièces', sections: 'Sections', results: '{n} résultats', truncated: 'Plus de {n} résultats : affinez la recherche.',
    none: 'Aucun résultat.', scoped: 'Limité au véhicule actif',
  },
  cart: {
    title: 'Panier', empty: 'Le panier est vide.', add: 'Ajouter au panier', added: 'Ajouté au panier', remove: 'Retirer',
    copy: 'Copier pour un tableur', copied: 'Copié — collez-le dans un tableur', csv: 'Exporter en CSV', print: 'Imprimer / PDF',
    share: 'Copier le lien de partage', linkCopied: 'Lien copié', shared: 'Panier partagé (lecture seule)', replace: 'Remplacer mon panier',
    merge: 'Ajouter à mon panier', missing: '{n} ligne(s) introuvable(s)', invalid: '{n} élément(s) invalide(s) ignoré(s)', open: 'Ouvrir le panier',
    section: 'Section', vehicle: 'Véhicule', qty: 'Qté', clear: 'Vider le panier',
  },
}
