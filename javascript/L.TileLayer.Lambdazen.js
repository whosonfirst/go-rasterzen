L.TileLayer.Lambdazen = L.TileLayer.extend({

    createTile: function (coords) {

	var img = new Image();
	img.setAttribute("height", 512);
	img.setAttribute("width", 512);	
	
	var url = this.getTileUrl(coords);
		
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
	    img.src = URL.createObjectURL(blob);
	};
	
	req.send();
	return img;
    }
    
});
