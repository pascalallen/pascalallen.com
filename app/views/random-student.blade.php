@extends('layouts.master')

@section('top-script')

	<style type="text/css">


		.static {
			position: absolute;
			width: 100%;
		}

		.instruct {
			text-align: center;
		    top: 250px;
		    left: 250px;
			position: absolute;
			font-size: 100px;
			color: white;
			text-shadow: -1px 0 black, 0 1px black, 1px 0 black, 0 -1px black, 10px 25px 5px white;
			display: none;
		}

		.student{
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
	    <canvas id="canvas"></canvas>
	    <span class="student"></span>
	    <h1 class="instruct">Calculating...</h1>
	    <audio id="contra" src="/data/contra.wav" type="audio/wav"></audio>

@stop

@section('bottom-script')

	<script type="text/javascript">

		"use strict";

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
		};

		$(document).keydown(function(e) {
			if(e.keyCode == 83){
				$('.instruct').fadeIn();
				setTimeout(function(event){
					$('.instruct').fadeOut();
					$('.static').fadeOut();
					//canvas init
					var canvas = document.getElementById("canvas");
					var ctx = canvas.getContext("2d");
					
					//canvas dimensions
					var W = window.innerWidth;
					var H = window.innerHeight;
					canvas.width = W;
					canvas.height = H;
					
					//snowflake particles
					var mp = 25; //max particles
					var particles = [];
					for(var i = 0; i < mp; i++)
					{
						particles.push({
							x: Math.random()*W, //x-coordinate
							y: Math.random()*H, //y-coordinate
							r: Math.random()*4+1, //radius
							d: Math.random()*mp, //density
				            color: "rgba(" + Math.floor((Math.random() * 255)) +", " + Math.floor((Math.random() * 255)) +", " + Math.floor((Math.random() * 255)) + ", 0.8)"
						})
					}
					
					//Lets draw the flakes
					function draw()
					{
						ctx.clearRect(0, 0, W, H);
						
						
						
						for(var i = 0; i < mp; i++)
						{ 
							var p = particles[i];
				            ctx.beginPath();
				            ctx.fillStyle = p.color;
							ctx.moveTo(p.x, p.y);
							ctx.arc(p.x, p.y, p.r, 0, Math.PI*2, true);
				            ctx.fill();
						}
						
						update();
					}
					
					//Function to move the snowflakes
					//angle will be an ongoing incremental flag. Sin and Cos functions will be applied to it to create vertical and horizontal movements of the flakes
					var angle = 0;
					function update()
					{
						angle += 0.01;
						for(var i = 0; i < mp; i++)
						{
							var p = particles[i];
							//Updating X and Y coordinates
							//We will add 1 to the cos function to prevent negative values which will lead flakes to move upwards
							//Every particle has its own density which can be used to make the downward movement different for each flake
							//Lets make it more random by adding in the radius
							p.y += Math.cos(angle+p.d) + 1 + p.r/2;
							p.x += Math.sin(angle) * 2;
							
							//Sending flakes back from the top when it exits
							//Lets make it a bit more organic and let flakes enter from the left and right also.
							if(p.x > W+5 || p.x < -5 || p.y > H)
							{
								if(i%3 > 0) //66.67% of the flakes
								{
				                    particles[i] = {x: Math.random()*W, y: -10, r: p.r, d: p.d, color : p.color};
								}
								else
								{
									//If the flake is exitting from the right
									if(Math.sin(angle) > 0)
									{
										//Enter from the left
				                        particles[i] = {x: -5, y: Math.random()*H, r: p.r, d: p.d, color: p.color};
									}
									else
									{
										//Enter from the right
				                        particles[i] = {x: W+5, y: Math.random()*H, r: p.r, d: p.d, color : p.color};
									}
								}
							}
						}
					}
					
					//animation loop
					setInterval(draw, 33);
					$('.student').text(getStudent());
	    			$('#contra').get(0).play();
				}, 3000);
			}
		});

	</script>

@stop