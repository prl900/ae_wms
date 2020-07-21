var WPS_URL='http://35.244.111.168:8080/wps';
var geojson=null;

var map = L.map('map', {
    zoom: 8,
    center: [-37., 145.],
});

// FeatureGroup is to store editable layers
var drawnItems = new L.FeatureGroup();
map.addLayer(drawnItems);
var drawControl = new L.Control.Draw({
    edit: {
        featureGroup: drawnItems
    }
});
map.addControl(drawControl);
map.on(L.Draw.Event.CREATED, function (event) {
    console.log(event);

    var layer = event.layer;
    drawnItems.clearLayers();
    drawnItems.addLayer(layer);
    geojson = layer.toGeoJSON().geometry;
    $.post(
        WPS_URL,
        JSON.stringify(geojson),
        function( result ) {
            console.log(result);
            var $table = $('<table>');
            $table.append('<thead>').children('thead')
                .append('<tr />').children('tr').append('<th>Year</th><th>Value</th>');
            var $tbody = $table.append('<tbody>').children('tbody');
            var lines = result.split('\n');
            lines.forEach(function(ln){
                if(!ln){
                    return;
                }
                var columns = ln.split(',');
                $tbody.append('<tr />').children('tr').last().append(columns.map(function(c){
                    return '<td>'+c+'</td>';
                }));
            });
            $('#timeseries').children().remove();
            $table.appendTo('#timeseries');

        });
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
var array = (new Array(20)).fill(0).map(function(_,ix){return 2001+ix;});

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

var TimeseriesTableControl = L.Control.extend({
    onAdd: function(map) {
        var dd = L.DomUtil.create('div');
        dd.id="timeseries";
        return dd;
    },

    onRemove: function(map) {
        // Nothing to do here
    }
});

var tsControl = function(opts){
    return new TimeseriesTableControl(opts);
}

deaLayer.addTo(map);
var yc = yearControl({
    position:'topright'
});
yc.addTo(map);

var tsc = tsControl({
    position:'bottomright'
});
tsc.addTo(map);

//var deaTimeLayer = L.timeDimension.layer.wms(deaLayer, {cache:0, cacheForward:0, cacheBackward:0});
//deaTimeLayer.addTo(map);
