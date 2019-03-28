from flask import Flask
from flask import request
from flask import send_file
from flask import render_template
from flask import make_response

import numpy as np
import numexpr as ne
import tensorflow as tf
import urllib
import os
import io
import matplotlib.pyplot as plt

tile_server = "https://geoarray-dot-wald-1526877012527.appspot.com/geoarray"

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

    sess.close()

    #return the result
    return ppx[0]


def get_tile(bbox, x_size, y_size, band, srs):
    contents = urllib.request.urlopen("{}?height={}&width={}&band={}&bbox={},{},{},{}&srs={}".format(tile_server, y_size, x_size, band, bbox[0], bbox[1], bbox[2], bbox[3], srs)).read()
    return np.frombuffer(contents, dtype=np.uint8).reshape((y_size, x_size)).astype(np.float32)


# If `entrypoint` is not defined in app.yaml, App Engine will look for an app
# called `app` in `main.py`.
app = Flask(__name__)

@app.route('/wms')
def wms():
    service = request.args.get('service')
    if service != 'WMS':
        return "Malformed request: only WMS requests implemented", 400

    req = request.args.get('request')
    if req == 'GetCapabilities':
        layers = [{'name': 'NDVI', 'title': 'Dynamic NDVI', 'abstract': 'AI generated'}]
        template = render_template('GetCapabilities.xml', layers=layers)
        response = make_response(template)
        response.headers['Content-Type'] = 'application/xml'
        return response

    if req != 'GetMap':
        return "Malformed request: only GetMap and GetCapabilities requests implemented", 400

    bbox = request.args.get('bbox').split(',')
    if len(bbox) != 4:
        return "Malformed request: bbox must have 4 values", 400

    #layer = request.args.get('layer')
    
    width = int(request.args.get('width'))
    height = int(request.args.get('height'))
    srs = request.args.get('srs')
    print(srs)
    styles = request.args.get('styles')

    nir = get_tile(bbox, width, height, 2, srs)
    red = get_tile(bbox, width, height, 1, srs)

    """
    ndvi = "(nir - red) / (nir + red)"
    res = ne.evaluate(ndvi)
    """
    
    res = tf_ndvi(red, nir)
    out = io.BytesIO()
    plt.imsave(out, res, cmap=styles, format="png")
    out.seek(0)

    return send_file(out, attachment_filename='tile.png', mimetype='image/png')


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=os.environ['PORT'], debug=True)
