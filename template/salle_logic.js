function getURLParam(name) {
    var url = window.location.search;
    var params = new URLSearchParams(url);
    return params.get(name);
}

var codeSalle = getURLParam('code');

if (codeSalle) {
    console.log("Code de la salle: " + codeSalle);
} else {
    console.log("Pas de code");
}

function actualiserListe() {
    fetch('/api/joueurs?code=' + codeSalle)
        .then(function(response) {
            return response.json();
        })
        .then(function(joueurs) {
            var liste = document.getElementById('listeJoueurs');
            liste.innerHTML = '';
            for (var i = 0; i < joueurs.length; i++) {
                var li = document.createElement('li');
                li.innerText = joueurs[i];
                liste.appendChild(li);
            }
        })
        .catch(function(error) {
            console.log('Erreur:', error);
        });
}

setInterval(actualiserListe, 2000);
actualiserListe();