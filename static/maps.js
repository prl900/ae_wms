var map = L.map('map', {
    zoom: 12,
    //center: [-37., 145.],
    center: [-29.46, 149.83],
    //timeDimension: true,
    /*timeDimensionOptions: {
        timeInterval: "2001-01-01/2010-01-01",
        period: "P1Y"
    },*/
    //timeDimensionControl: true,
});

//L.tileLayer('http://{s}.tile.osm.org/{z}/{x}/{y}.png', {
//L.tileLayer('http://tiles.wmflabs.org/hillshading/{z}/{x}/{y}.png', {
L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Shaded_Relief/MapServer/tile/{z}/{y}/{x}', {
    attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a>'
}).addTo(map);

//var wmsUrl = "https://dea-wms-dot-wald-1526877012527.appspot.com/wms"
var wmsUrl = "http://35.244.111.168:8080/wms"

var deaLayer = L.tileLayer.wms(wmsUrl, {
    layers: 'wcf',
    format: 'image/png',
    opacity: 1.0,
    transparent: true,
    updateWhenIdle: true,
    updateWhenZooming: false,
    updateInterval: 500,
    attribution: '<a href="http://wald.anu.edu.au/">WALD ANU</a>'
})
	
deaLayer.addTo(map);

//var deaTimeLayer = L.timeDimension.layer.wms(deaLayer);
//deaTimeLayer.addTo(map);

