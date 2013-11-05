var page = require('webpage').create(),
    system = require('system'),
    fs = require('fs'), address;


if (system.args.length < 2) {
	console.log('Usage: download.js <URL> <UA> <Timeout>');
	phantom.exit(1);
} else {
	page.onError = function(msg, trace) { /*console.log(msg);*/ };
	page.onResourceRequested = function(requestData, request) {
    		//console.log('Request (#' + requestData.id + '): ' + JSON.stringify(requestData));
		var url = requestData['url'];
		if (requestData['Content-Type'] != null && requestData['Content-Type'] !== "undefined")
		{
			var type = requestData['Content-Type'].toLowerCase();
			if (type.indexOf('audio') != -1 || type.indexOf('image') != -1 || type.indexOf('video') != -1)
			{
				//console.log("Abort");
				request.abort();
			}
		}
		if (url.indexOf('.jpg') != -1 || url.indexOf('.png') != -1 || url.indexOf('.bmp') != -1 || url.indexOf('.gif') != -1)
		{
			//console.log("Abort");
			request.abort();
		}
	};

	address = system.args[1];
	//UA Setting
	page.settings.userAgent = 'Mozilla/5.0 (Windows; U; Windows NT 5.1; en-US) AppleWebKit/525.13 (KHTML, like Gecko) Chrome/0.A.B.C Safari/525.13';
	if (system.args.length >= 3)
	{
		page.settings.userAgent = system.args[2];
	}	
	//Timeout
	page.settings.resourceTimeout = 10000; //Timeout in ms
	if (system.args.length >= 4)
	{
		page.settings.resourceTimeout = system.args[3];
	}	
	page.onResourceTimeout = function(msg) { phantom.exit(1); }
	page.onConsoleMessage = function(msg) { /*console.log(msg);*/ }    
	//open url
	page.open(address, function (status) {
		if (status !== 'success') {
			console.log('FAIL to load the address');
			phantom.exit(2);
		} else {
			console.log(page.content);
			//console.log(page.frameContent);
			phantom.exit();
		}
	});
}
