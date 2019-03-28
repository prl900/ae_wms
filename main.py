from flask import Flask
from flask import request
from flask import send_file

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


def get_tile(bbox, x_size, y_size, band):
    contents = urllib.request.urlopen("{}?height={}&width={}&band={}&bbox={},{},{},{}".format(tile_server, y_size, x_size, band, bbox[0], bbox[1], bbox[2], bbox[3])).read()
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
    if req != 'GetMap':
        return "Malformed request: only GetMap requests implemented", 400
    
    bbox = request.args.get('bbox').split(',')
    if len(bbox) != 4:
        return "Malformed request: bbox must have 4 values", 400

    width = request.args.get('width')
    height = request.args.get('height')
    srs = request.args.get('srs')

    bbox = [0.0, 0.0, 10.0, 10.0]
    x_size = 256
    y_size = 256
    nir = get_tile(bbox, x_size, y_size, 2)
    red = get_tile(bbox, x_size, y_size, 1)

    """
    ndvi = "(nir - red) / (nir + red)"
    res = ne.evaluate(ndvi)
    """

    res = tf_ndvi(red, nir)
    out = io.BytesIO()
    plt.imsave(out, res, cmap="summer_r", format="png")
    out.seek(0)
   
    return send_file(out, attachment_filename='tile.png', mimetype='image/png')


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=os.environ['PORT'], debug=True)
