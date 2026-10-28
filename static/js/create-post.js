// Récupérer les cases à cocher du groupe des catégories.
const casesCategories = document.querySelectorAll(
    '#categories-group input[type="checkbox"][name="categories"]'
);

const maximumCategories = 5;

function mettreAJourCategories() {
    // Éviter d'accéder à une première case qui n'existerait pas.
    if (casesCategories.length === 0) {
        return;
    }

    // Compter les cases actuellement cochées.
    let nombreCochees = 0;

    for (const caseCategorie of casesCategories) {
        if (caseCategorie.checked) {
            nombreCochees++;
        }
    }

    // Exiger au moins une catégorie.
    if (nombreCochees === 0) {
        casesCategories[0].setCustomValidity(
            "Sélectionne au moins une technologie."
        );
    } else {
        casesCategories[0].setCustomValidity("");
    }

    // À cinq choix, désactiver uniquement les cases non cochées.
    for (const caseCategorie of casesCategories) {
        if (nombreCochees >= maximumCategories && !caseCategorie.checked) {
            caseCategorie.disabled = true;
        } else {
            caseCategorie.disabled = false;
        }
    }
}

// Refaire les contrôles après chaque changement de sélection.
for (const caseCategorie of casesCategories) {
    caseCategorie.addEventListener("change", mettreAJourCategories);
}

// Appliquer les contrôles dès l'ouverture de la page.
mettreAJourCategories();