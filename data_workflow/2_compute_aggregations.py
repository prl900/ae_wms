import numpy as np
from glob import glob
from osgeo import gdal
import os

var_name = "maxPV"

def compose_tile(i, j, level):
    step = 2**level
    tile = np.ones((4000,4000), dtype=np.uint8)*255

    for i_n in range(step):
        for j_n in range(step):
            fname = "/g/data/ub8/au/blobs/fc_metrics_maxPV_{:+03d}_{:+03d}_l0_2001.npy".format(i+i_n, j-j_n)
            if os.path.isfile(fname):
                chunk = np.load(fname)
                tile[4000*j_n//step:4000*(j_n+1)//step, 4000*i_n//step:4000*(i_n+1)//step] = chunk[::step, ::step]

    return tile


def create_aggregation(level):
    step = 2**level
    for i in range(-19, 21+1, step):
        for j in range(-10, -48-1, -1*step):
            print(i,j)
            tile = compose_tile(i,j,level)
            np.save("/g/data/ub8/au/blobs/fc_metrics_maxPV_{:+03d}_{:+03d}_l{}_2001.npy".format(i, j, level), tile)


#create_aggregation(1)
#create_aggregation(2)
create_aggregation(3)
create_aggregation(4)
create_aggregation(5)

