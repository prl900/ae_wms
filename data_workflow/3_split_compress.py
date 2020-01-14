import numpy as np
from glob import glob
from osgeo import gdal
import os
import snappy

var_name = "maxPV"

def save_chunks(i, j, level):

    fname = "/g/data/ub8/au/blobs/fc_metrics_maxPV_{:+03d}_{:+03d}_l{:}_2001.npy".format(i, j, level)
    if not os.path.isfile(fname):
        return

    step = 2**level
    tile = np.load(fname)

    chunks = {}
    for ci in range(10):
        for cj in range(10):
            chunk = tile[cj*400:(cj+1)*400, ci*400:(ci+1)*400]
            if np.any(chunk != 255):
                chunks["{:+04d}_{:+04d}".format(i*10+ci*step,j*10-cj*step)] = chunk
                data = chunk.tobytes()
                with open("/g/data/ub8/au/blobs/fc_metrics_maxPV_{:+04d}_{:+04d}_l{}_2001.snp".format(i*10+ci*step,j*10-cj*step,level), 'wb') as out_file:
                    compressed = snappy.compress(data)
                    out_file.write(compressed)


    if chunks:
        np.savez(fname[:-4], **chunks)


def chunk_tiles(level):
    step = 2**level
    for i in range(-19, 21+1, step):
        for j in range(-10, -48-1, -1*step):
            save_chunks(i, j, level)


#chunk_tiles(0)
#chunk_tiles(1)
#chunk_tiles(2)
chunk_tiles(3)
chunk_tiles(4)
chunk_tiles(5)
