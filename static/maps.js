var map = L.map('map', {
    zoom: 8,
    center: [-37., 145.],
});

L.tileLayer('https://server.arcgisonline.com/ArcGIS/rest/services/World_Shaded_Relief/MapServer/tile/{z}/{y}/{x}', {
    attribution: '&copy; <a href="http://osm.org/copyright">OpenStreetMap</a>'
}).addTo(map);

//var deaLayer = L.tileLayer.wms("http://35.244.111.168:8080/wms", {
var deaLayer = L.tileLayer.wms("http://localhost:8080/wms", {
    layers: 'kc',
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
