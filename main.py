from flask import Flask
from flask import request
from flask import send_file
from osgeo import gdal
import numpy as np
import numexpr as ne
from google.cloud import storage
import io
import matplotlib.pyplot as plt

# If `entrypoint` is not defined in app.yaml, App Engine will look for an app
# called `app` in `main.py`.
app = Flask(__name__)

@app.route('/')
def hello():
    #color_map = request.args.get('cmap')
    expr = str(request.query_string).strip("'").split('=')[1]

    client = storage.Client()
    # https://console.cloud.google.com/storage/browser/[bucket-id]/
    bucket = client.get_bucket('wald-1526877012527.appspot.com')
    # Then do other things...
    blob = bucket.get_blob('red_lr.npy')
    f = io.BytesIO(blob.download_as_string())
    red = np.load(f)
    
    blob = bucket.get_blob('nir_lr.npy')
    f = io.BytesIO(blob.download_as_string())
    nir = np.load(f)
    f = None

    #ndvi = (nir - red) / (nir + red)
    res = ne.evaluate(expr)
    red = None
    nir = None
    
    out = io.BytesIO()
    #plt.imsave(out, res, cmap=color_map, format="png")
    plt.imsave(out, res, cmap="summer_r", format="png")
    out.seek(0)
    return send_file(out, attachment_filename='tile.png', mimetype='image/png')
    #a = np.ones((10,10))
    #b = np.arange(100).reshape((10,10))


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=8080, debug=True)
