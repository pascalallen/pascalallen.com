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

		.arrow-right {
			width: 0; 
			height: 0; 
			border-top: 10px solid transparent;
			border-bottom: 10px solid transparent;
			border-left: 10px solid black;
			left: 880px;
    		top: 405px;
			position: absolute;
		}

		.arrow-left {
			width: 0; 
			height: 0; 
			border-top: 10px solid transparent;
			border-bottom: 10px solid transparent; 
			border-right:10px solid black;
			left: 845px;
			top: 405px;
			position: absolute;
		}

	</style>

@stop

@section('content')

    <img class="static" src="img/static.gif">
    <span class="contra"></span>
    <span class="slideOne"></span>
    <img class="tv" src="img/tv.png">
    <h1 class="title">You made it.</h1>
    <audio id="contra" src="/data/contra.wav" type="audio/wav"></audio>
	<div class="arrow-left"></div>
	<div class="arrow-right"></div>

@stop

@section('bottom-script')

<script type="text/javascript">

	"use strict";

	$(document).ready(setTimeout(function(event){
    	$('.contra').css('background-image', 'url(/img/contra.gif)').css('background-size', '100%').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px').css('position', 'absolute');
    	$('#contra').get(0).play();
	}, 2000));

	$(document).keydown(function(e) {
		if(e.keyCode == 37){
			console.log("left");
    		$('.slideOne').css('background-image', 'url(http://placehold.it/350x150)').css('background-size', '100%').css('top', '175px').css('left', '320px').css('width', '450px').css('height', '375px').css('position', 'absolute');
		}
		if(e.keyCode == 39) {
			console.log("right");
		}
    });

</script>

@stop
