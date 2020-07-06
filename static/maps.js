var map = L.map('map', {
    zoom: 8,
    center: [-37., 145.],
    //center: [-29.46, 149.83],
    //timeDimension: true,
    /*timeDimensionOptions: {
        timeInterval: "2001-01-01/2003-01-01",
        period: "P1Y"
    },
    timeDimensionControl: true,*/
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
var array = (new Array(10)).fill(0).map(function(_,ix){return 2001+ix;});

var yearChanged = function(e){
    console.log(e.target.value);
    deaLayer.setParams({
        time:e.target.value+'-01-01T00:00:00.000Z'
    });
};

var YearControl = L.Control.extend({
    onAdd: function(map) {
        var dd = L.DomUtil.create('select');

        dd.style.width = '200px';

        for (var i = 0; i < array.length; i++) {
            var option = document.createElement("option");
            option.value = array[i];
            option.text = array[i].toString();
            dd.appendChild(option);
        }
        dd.addEventListener('input', yearChanged);
        return dd;
    },

    onRemove: function(map) {
        // Nothing to do here
    }
});

var yearControl = function(opts){
    return new YearControl(opts);
};

deaLayer.addTo(map);
var yc = yearControl({
    position:'topright'
});
yc.addTo(map);

//var deaTimeLayer = L.timeDimension.layer.wms(deaLayer, {cache:0, cacheForward:0, cacheBackward:0});
//deaTimeLayer.addTo(map);
