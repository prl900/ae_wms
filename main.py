from flask import Flask
from flask import request as freq
from flask import send_file
import numpy as np
import numexpr as ne
import tensorflow as tf
import urllib

import io
import matplotlib.pyplot as plt

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


# If `entrypoint` is not defined in app.yaml, App Engine will look for an app
# called `app` in `main.py`.
app = Flask(__name__)

@app.route('/')
def hello():
    contents = urllib.request.urlopen("https://geoarray-dot-wald-1526877012527.appspot.com/geoarray?height=256&width=256&band=2&bbox=0.0,0.0,10.0,10.0").read()
    nir = np.frombuffer(contents, dtype=np.uint8).reshape((256,256)).astype(np.float32)

    contents = urllib.request.urlopen("https://geoarray-dot-wald-1526877012527.appspot.com/geoarray?height=256&width=256&band=1&bbox=0.0,0.0,10.0,10.0").read()
    red = np.frombuffer(contents, dtype=np.uint8).reshape((256,256)).astype(np.float32)


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
    app.run(host='127.0.0.1', port=8080, debug=True)
