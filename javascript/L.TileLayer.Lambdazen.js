L.TileLayer.Lambdazen = L.TileLayer.extend({

    createTile: function (coords) {

	var url = this.getTileUrl(coords);

	var tile = L.DomUtil.create('canvas', 'leaflet-tile');
        tile.width = 512;
        tile.height = 512;
		
	var req = new XMLHttpRequest();	
	req.open("GET", url, true);

	// if you're wondering this entire package exists for the
	// sole purpose of being able to send this request header
	// because this: https://github.com/whosonfirst/go-rasterzen#lambda-api-gateway-and-images
	// (20180713/thisisaaronland)
	
	req.setRequestHeader("Accept", "image/png");
	
	req.responseType = "blob";
	
	req.onerror = function(rsp){
	    console.log("SAD", url, rsp);
	};
	
	req.onload = function(rsp){

	    var blob = req.response;	    
	    var img = new Image();
	    
	    img.onload = function(){
		var context = tile.getContext('2d');		
		context.drawImage(img, 0, 0);
	    };

	    img.src = URL.createObjectURL(blob);
	};
	
	req.send();
	return tile;
    }
    
});
