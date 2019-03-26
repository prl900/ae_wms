from flask import Flask
from flask import request as freq
from flask import send_file
import numpy as np
import numexpr as ne
import urllib

import io
import matplotlib.pyplot as plt


# If `entrypoint` is not defined in app.yaml, App Engine will look for an app
# called `app` in `main.py`.
app = Flask(__name__)

@app.route('/')
def hello():
    contents = urllib.request.urlopen("https://geoarray-dot-wald-1526877012527.appspot.com/geoarray?height=256&width=256&band=2&bbox=0.0,0.0,10.0,10.0").read()
    nir = np.frombuffer(contents, dtype=np.uint8).reshape((256,256)).astype(np.float32)

    contents = urllib.request.urlopen("https://geoarray-dot-wald-1526877012527.appspot.com/geoarray?height=256&width=256&band=1&bbox=0.0,0.0,10.0,10.0").read()
    red = np.frombuffer(contents, dtype=np.uint8).reshape((256,256)).astype(np.float32)


    ndvi = "(nir - red) / (nir + red)"
    res = ne.evaluate(ndvi)
    
    out = io.BytesIO()
    plt.imsave(out, res, cmap="summer_r", format="png")
    out.seek(0)
    
    return send_file(out, attachment_filename='tile.png', mimetype='image/png')


if __name__ == '__main__':
    app.run(host='127.0.0.1', port=8080, debug=True)
