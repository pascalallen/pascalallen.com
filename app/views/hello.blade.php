@extends('layouts.master')

@section('top-script')

	<style>

		.title {
			text-align: center;
		    top: 255px;
			left: 380.75px;
			position: absolute;
			font-size: 60px;
			color: white;
			text-shadow: -1px 0 black, 0 1px black, 1px 0 black, 0 -1px black, 10px 25px 5px white;
		}

		.instruct {
			text-align: center;
		    top: 250px;
			left: 380.75px;
			position: absolute;
			font-size: 60px;
			color: white;
			text-shadow: -1px 0 black, 0 1px black, 1px 0 black, 0 -1px black, 10px 25px 5px white;
			display: none;
		}

		body {
			background-image: url("/img/grass_texture.png");
			height: 100%;
			width: 100%;
			background-color: rgba(44, 176, 55, 0.3);
			/*position: absolute;*/
			top: 0;
		}

		.tv {
			position:absolute;
			top:140px;
			left:250px;
			width:711px;
			height:450px;
		}

		.static {
			position:absolute;
			top: 175px;
    		left: 320px;
			width:450px;
			height:375px;
		}

		.slide{
			font-size: 100px;
			position: absolute;
			text-align: center;
			top: 260px;
			left: 450px;
		}

	</style>

@stop

@section('content')

    <img class="static" src="img/static.gif">
    <span class="white"></span>
    {{-- <span class="contra"></span> --}}
    <span class="slide"></span>
    <img class="tv" src="img/tv.png">
    {{-- <h1 class="title">You made it.</h1> --}}
    <h1 class="instruct">Calculating...</h1>
    <audio id="contra" src="/data/contra.wav" type="audio/wav"></audio>
@stop

@section('bottom-script')

<script type="text/javascript">

	"use strict";

	// $(document).ready(setTimeout(function(event){
    	// $('.contra').css('background-image', 'url(/img/contra.gif)').css('background-size', '100%').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px').css('position', 'absolute');
    	// $('#contra').get(0).play();
	// }, 2000));

	// var array = ["/img/pic1.jpg", "/img/pic2.jpg", "/img/pic3.jpg", "/img/pic4.jpg", "/img/pic5.jpg"];
	// var count = -1;
	// $(document).keydown(function(e) {
	// 	if(e.keyCode == 37 && count != 0){
	// 		console.log("left");
 //    		count--;
 //    		$('.slide').css('background-image', 'url(' + array[count] + ')').css('background-size', '100%').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px').css('position', 'absolute').attr("href", "https://www.google.com/");
	// 	}
	// 	if(e.keyCode == 39 && count != 4){
	// 		console.log("right");
 //    		count++;
 //    		$('.slide').css('background-image', 'url(' + array[count] + ')').css('background-size', '100%').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px').css('position', 'absolute');
	//  	}

 //    });

 //    $(document).ready(function(e){
 //    	$('.title').fadeOut(2000);
 //    	$('.instruct').fadeIn(2000);
 //    	$('.instruct').fadeOut(2000);
 //    });

    var students = [
    	"Joseph", 
    	"Greg", 
    	"Burney", 
    	"Gabe", 
    	"Jean", 
    	"Richard", 
    	"Logan", 
    	"Stan", 
    	"Tomas", 
    	"Gaston", 
    	"Jim", 
    	"Margot", 
    	"Don", 
    	"Anna", 
    	"Rick", 
    	"Dan", 
    	"CJ", 
    	"Trey", 
    	"Nick"
	];

	function getStudent() {
	   return students[Math.floor(Math.random() * students.length)];
	}

	$(document).keydown(function(e) {
		if(e.keyCode == 83){
			$('.instruct').fadeIn();
			setTimeout(function(event){
				$('.instruct').fadeOut();
				$('.white').css('background-color', 'white').css('position', 'absolute').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px');
				$('.slide').text(getStudent());
    			$('#contra').get(0).play();
			}, 3000);
    		// $('.slide').css('background-image', 'url(' + array[count] + ')').css('background-size', '100%').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px').css('position', 'absolute');
		}
	});

	// .css('font-size', '60px').css('position', 'absolute').css('text-align', 'center')

</script>

@stop
