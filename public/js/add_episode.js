function isInt(value) {
  return !isNaN(value) && 
         parseInt(Number(value)) == value && 
         !isNaN(parseInt(value, 10));
}


function flash(msg, type){
	var flashmsg= $("#flashmsg");
	flashmsg.html(msg);
	flashmsg.removeClass();
	flashmsg.addClass("alert alert-"+type);
	flashmsg.show();
}

$(function() {
	$("#flashmsg").hide();

    $("#formAddEp").submit(function(event) {
    	event.preventDefault();
    	var podcast = $("#podcast").val();
    	if (podcast.length==0) {
    		flash("Please fill podcats field", "danger");
    		return;
    	}
    	var episode = $("#episode").val();
    	//if (isNaN(episode) || (parseInt(episode) != episode)) {
    	if (!isInt(episode)) {
    		flash("Episode must be an integer", "danger");
    		return;
    	};    	
    	var ptitle = $("#title").val();
    	if (ptitle.length==0) {
    		flash("Please fill title field", "danger");
    		return;
    	}    	
    	var link = $("#link").val();
    	if (link.length==0) {
    		flash("Please fill title link", "danger");
    		return;
    	}      	
    	var counter_diff =  $("#counter_diff").val();
    	if (!isInt(counter_diff)){
    		counter_diff=0;
    	}

        var data=`{"podcast":"`+ podcast+`","episode":`+episode+`,"title":"`+ptitle+`","link":"`+ link+`","counter_diff":`+counter_diff+`}`;
        console.log(data);
        
        $.ajax({
            url: "/a/add",
            type: "POST",
            data: data,
            dataType: 'json',
            success: function(json) {
                console.log(json);
                json.success ? flash(json.msg,"success"): flash(json.msg,"danger");  
            },
            error: function(xhr, status, errorThrown) {
            	flash("Internal server error","danger");
                console.log("Error: " + errorThrown);
                console.log("Status: " + status);
                console.dir(xhr);
            },
        });
        
    })

});
