from flask import Flask
from flask import request
from flask import send_file
from flask import render_template
from flask import make_response

#from tensorly import tucker_to_tensor
from pyproj import Proj, transform
import imageio
import numpy as np
import math
import numexpr as ne
#import tensorflow as tf
import urllib
import os
import io
import matplotlib.pyplot as plt
from google.cloud import storage
import time

"""
import googlecloudprofiler

# Profiler initialization. It starts a daemon thread which continuously
# collects and uploads profiles. Best done as early as possible.
try:
    # service and service_version can be automatically inferred when
    # running on App Engine. project_id must be set if not running
    # on GCP.
    googlecloudprofiler.start(verbose=3)
except (ValueError, NotImplementedError) as exc:
    print(exc)  # Handle errors here
"""



sinu_proj = "+proj=sinu +lon_0=0 +x_0=0 +y_0=0 +a=6371007.181 +b=6371007.181 +units=m +no_defs "
wgs84_proj = "epsg:4326"
modis_pixel_size = (463.312716527916507, 463.312716527916677)
modis_tile = "MCD43A4.A2018001.h%02dv%02d.006_b%d_16"
modis_tile_size = 2400

mylog = "Welcome to the home made profiling stack:<br>"
    
storage_client = storage.Client()
bucket = storage_client.get_bucket('tiny_map')

def xy2ij(origin, pixel_size, x, y):
    i = round((x - origin[0]) / pixel_size[0])
    j = round((origin[1] - y) / pixel_size[1])
    
    return i, j
    
    
def get_partial_tile(bbox, b, h, v, im_size=256, proj=wgs84_proj):
    global mylog
    # bbox contains [min_lon, min_lat, max_lon, max_lat]
    pixel_size = ((bbox[2] - bbox[0]) / im_size, (bbox[3] - bbox[1]) / im_size)

    arr = np.zeros((im_size,im_size), dtype=np.uint16)
    
    start = time.time()
    lons = []
    lats = []
    for j, lat in enumerate(np.arange(bbox[3], bbox[1], -pixel_size[1])):
        for i, lon in enumerate(np.arange(bbox[0], bbox[2], pixel_size[0])):
            lons.append(lon)
            lats.append(lat)
    
    inProj = Proj(init=proj)
    outProj = Proj(sinu_proj)
    xs, ys = transform(inProj, outProj, lons, lats)
    mylog += "    {}: calculating coordinates<br>".format(time.time() - start)
   
    # Instantiates a client
    start = time.time()
    blob = bucket.blob(modis_tile % (h, v, b))
    mylog += "       File: {}<br>".format(modis_tile % (h, v, b))
    mylog += "       {}: setting up storage<br>".format(time.time() - start)
    try:
        f = io.BytesIO(blob.download_as_string())
        f.seek(0)
        mylog += "       {}: downloading image<br>".format(time.time() - start)
    except:
        return arr

    im = imageio.imread(f)
    mylog += "       {}: reading image<br>".format(time.time() - start)
    mylog += "    {}: total image<br>".format(time.time() - start)
    
    start = time.time()
    origin = ((h-18)*modis_pixel_size[0]*modis_tile_size, (9-v)*modis_pixel_size[1]*modis_tile_size)
    for j in range(im_size):
        for i in range(im_size):
            oi, oj = xy2ij(origin, modis_pixel_size, xs[j*im_size+i], ys[j*im_size+i])

            if oi > modis_tile_size - 1 or oj > modis_tile_size - 1:
                arr[j,i] = 41248
                continue
            if oi < 0 or oj < 0:
                arr[j,i] = 41248
                continue
            arr[j,i] = im[oj,oi]  
    mylog += "    {}: reprojecting image<br>".format(time.time() - start)
            
    return arr

def bbox2tile(bbox, band, im_size, proj):
    global mylog
    pixel_size = (bbox[2] - bbox[0]) / im_size
    modis_x_extent = modis_pixel_size[0]*modis_tile_size
    modis_y_extent = modis_pixel_size[1]*modis_tile_size
    inProj = Proj(init=proj)
    outProj = Proj(sinu_proj)
 
    x_tl, y_tl = transform(inProj, outProj, bbox[0], bbox[3])
    x_tr, y_tr = transform(inProj, outProj, bbox[2], bbox[3])
    x_br, y_br = transform(inProj, outProj, bbox[2], bbox[1])
    x_bl, y_bl = transform(inProj, outProj, bbox[0], bbox[1])
    
    max_h = max(int(math.floor(x_tl/modis_x_extent)), int(math.floor(x_tr/modis_x_extent))) + 18
    min_h = min(int(math.floor(x_br/modis_x_extent)), int(math.floor(x_bl/modis_x_extent))) + 18
    min_v = min(-1*int(math.ceil(y_tl/modis_y_extent)), -1*int(math.ceil(y_bl/modis_y_extent))) + 9
    max_v = max(-1*int(math.ceil(y_br/modis_y_extent)), -1*int(math.ceil(y_tr/modis_y_extent))) + 9
    
    arr = None
    for h in range(min_h, max_h+1):
        #for v in range(min_v, max_v+1):
        for v in range(max_v, min_v-1, -1):
            mylog += " {}, {}: transform tile<br>".format(h, v)
            #return get_partial_tile(bbox, band, h, v, im_size, proj)
            a = get_partial_tile(bbox, band, h, v, im_size, proj)
            a[a == 41248] = 0
            if arr is None:
                arr = a
                continue
                
            arr += a
    
    return arr

"""
def tf_ndvi(red, nir):
    r = tf.placeholder(tf.float32)
    n = tf.placeholder(tf.float32)

    #build the ndvi operation
    ndvi = (n - r) / (n + r)

    #get the tensorflow session
    sess = tf.Session()

    #initialize all variables
    sess.run(tf.initialize_all_variables())

    feed_dict = {r:red, n:nir}

    #now run the sum operation
    ppx = sess.run([ndvi], feed_dict)
    ndvi = None
    sess.close()

    #return the result
    return ppx[0]
"""

def get_tile(bbox, x_size, y_size, band, srs):
    contents = urllib.request.urlopen("{}?height={}&width={}&band={}&bbox={},{},{},{}&srs={}".format(tile_server, y_size, x_size, band, bbox[0], bbox[1], bbox[2], bbox[3], srs)).read()
    return np.frombuffer(contents, dtype=np.uint8).reshape((y_size, x_size)).astype(np.float32)


# If `entrypoint` is not defined in app.yaml, App Engine will look for an app
# called `app` in `main.py`.
app = Flask(__name__)

@app.route('/wms')
def wms():
    global mylog
    mylog = ''
    start = time.time()
    service = request.args.get('service')
    if service != 'WMS':
        return "Malformed request: only WMS requests implemented", 400

    req = request.args.get('request')
    if req == 'GetCapabilities':
        layers = [{'name': 'NDVI', 'title': 'Dynamic NDVI', 'abstract': 'AI generated'}]
        template = render_template('GetCapabilities.xml', layers=layers)
        response = make_response(template)
        response.headers['Content-Type'] = 'application/xml'
        response.headers['Access-Control-Allow-Origin'] = '*'
        return response

    if req != 'GetMap':
        return "Malformed request: only GetMap and GetCapabilities requests implemented", 400

    bbox = request.args.get('bbox').split(',')
    if len(bbox) != 4:
        return "Malformed request: bbox must have 4 values", 400
    bbox = [float(p) for p in bbox]

    #layer = request.args.get('layer')
    
    width = int(request.args.get('width'))
    height = int(request.args.get('height'))
    srs = request.args.get('srs').lower()

    #return "{} {} {} {}".format(bbox, width, height, srs)
    #styles = request.args.get('styles')
    #styles = "summer_r"
    styles = "RdYlGn"
    mylog += "{}: parsing fields<br>".format(time.time() - start)

    #nir = bbox2tile(bbox, 2, im_size, proj)
    #return bbox2tile(bbox, 1, width, srs)
    start = time.time()
    red = bbox2tile(bbox, 1, width, srs)
    nir = bbox2tile(bbox, 2, width, srs)
    mylog += "{}: creating tile<br>".format(time.time() - start)
    #return "{} {}".format(red.shape, red.dtype)
    #nir = get_tile(bbox, width, height, 2, srs)
    #red = get_tile(bbox, width, height, 1, srs)

    ndvi = "(nir - red) / (nir + red)"
    res = ne.evaluate(ndvi)
    
    #res = tf_ndvi(red, nir)
    #red = None
    #nir = None
    start = time.time()
    out = io.BytesIO()
    plt.imsave(out, res, cmap=styles, format="png")
    #res = None
    out.seek(0)
    mylog += "{}: encoding tile<br>".format(time.time() - start)
    #return mylog
    return send_file(out, attachment_filename='tile.png', mimetype='image/png')

"""
@app.route('/proj')
def proj():
    inProj = Proj(init='epsg:3857')
    outProj = Proj(init='epsg:4326')
    x1,y1 = -11705274.6374,4826473.6922
    x2,y2 = transform(inProj,outProj,x1,y1)
    return "{} {}".format(x2, y2)
"""

if __name__ == '__main__':
    app.run(host='127.0.0.1', port=os.environ['PORT'], debug=True)
