var map = L.map('map', {
    zoom: 8,
    center: [-21., 149.],
    //timeDimension: true,
    /*timeDimensionOptions: {
        timeInterval: "2001-01-01/2010-01-01",
        period: "P1Y"
    },*/
    //timeDimensionControl: true,
});

L.tileLayer('http://{s}.tile.osm.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a>'
}).addTo(map);

//var wmsUrl = "https://dea-wms-dot-wald-1526877012527.appspot.com/wms"
var wmsUrl = "https://dea-wms-dot-wald-1526877012527.appspot.com/wms"
              https://dea-wms-dot-wald-1526877012527.appspot.com/
var deaLayer = L.tileLayer.wms(wmsUrl, {
    layers: 'dea',
    format: 'image/png',
    opacity: 0.7,
    transparent: true,
    attribution: '<a href="http://wald.anu.edu.au/">WALD ANU</a>'
})
	
deaLayer.addTo(map);

//var deaTimeLayer = L.timeDimension.layer.wms(deaLayer);
//deaTimeLayer.addTo(map);

