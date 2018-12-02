# copernicus
REST API for querying of Sentinel-2 satellite data. The [Sentinel-2 satellite](https://en.wikipedia.org/wiki/Sentinel-2) is a part of the ESA Copernicus program to perform terrestrial observations for a wide range of purposes. The [Sentinel-2 dataset](https://cloud.google.com/storage/docs/public-datasets/sentinel-2) is preprocessed as a L-1C data product and hosted on Google Cloud: [gs://gcp-public-data-sentinel-2](https://console.cloud.google.com/storage/browser/gcp-public-data-sentinel-2/)

Retrieving satellite data for a specific location from the raw dataset is clumsy. This API offers easy querying based on either coordinates or an address. The service is hosted on Google App Engine.

*  **coordinates** 
   
   For a given set of coordinates, returns the paths of the blue, red and green of the three most recent images.
   Required:
   `Lon=float`,
   `Lat=float`
   
*  **morecoordinates** 
    
    For two given sets of coordinates, returns the paths of the blue, red and green of the three most recent images.        Required:
   `WestLon=float`,
   `EastLon=float`,
   `SouthLat=float`,
   `NorthLat=float`
   
*  **address** 

    For a given address, returns the paths of the blue, red and green of the three most recent images.
    Required:
    `Address=string`

*  **brank** 

    For a given address, returns the paths to the blue bands of the three most recent images, ranked in increasing order by  the distance to a colour value of 255.    
    Required:
    `Address=string`
    
*  **rgbrank** 

    Returns, for a given address, the paths to the three most recent images, ranked in increasing order by the distance to a given target hex encoded rgb colour.  
    Required:
    `Address=string`,
    `Color=string`



