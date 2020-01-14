import numpy as np
from glob import glob
import xarray as xr

def n_printer(n):
    if n >= 0:
        return "+{0:02d}".format(n)
    else:
        return "-{0:02d}".format(abs(n))

variables = ['maxPV', 'minPV']#, medPV, tmaxPV, medNPV, maxBS, tmaxBS, medwater, WCF]
var_name = variables[0]
paths = glob("/g/data/ub8/au/LandCover/DEA_ALC/*")

for path in paths:
    print(path)

    coords = path.split('/')[-1].split('_')
    coord_a = int(coords[0])
    coord_b = int(coords[1])

    # Add one to Y to desginate top-left instead of DEA bottom-left
    coord_str = "{:+03d}_{:+03d}".format(coord_a, coord_b+1)

    fname = path + "/fc_metrics_{}_2001.nc".format(path.split('/')[-1])
    ds = xr.open_dataset(fname)
    arr = ds[var_name].values.T
    arr[np.isnan(arr)] = 255
    arr = arr.astype(np.uint8)

    np.save("/g/data/ub8/au/blobs/fc_metrics_{}_{}_l0_2001".format(var_name, coord_str), arr)
