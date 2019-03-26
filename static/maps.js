var map = L.map('map', {
    zoom: 8,
    center: [-35.5, 149.],
});

L.tileLayer('http://{s}.tile.osm.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a>'
}).addTo(map);

var wmsUrl = "https://wald-1526877012527.appspot.com/wms"

var atmosLayer = L.tileLayer.wms(wmsUrl, {
    layers: 'e0',
    format: 'image/png',
    opacity: 0.7,
    transparent: true,
    attribution: '<a href="http://wald.anu.edu.au/">WALD ANU</a>'
}).addTo(map);
