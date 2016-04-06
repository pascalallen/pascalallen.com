@extends('layouts.master')

@section('top-script')

	<style>

		.title {
			text-align: center;
		    top: 200px;
			left: 380.75px;
			position: absolute;
			font-size: 60px;
			text-shadow: 10px 10px 15px #2CB037;
		}

		body {
			background-image: url("/img/grass_texture.png");
			height: 100%;
			width: 100%;
			position: relative;
			background-color: rgba(44, 176, 55, 0.3);
			height: 100%;
			width: 100%;
			position: absolute;
			top: 0;
		}

	</style>

@stop

@section('content')

    <img style="position:absolute; top:20%; left:27%; width:450px; height:375px" src="img/static.gif">
    <img style="position:absolute; top:15%; left:21%; width:711px; height:450px" src="img/tv.png">
    <h2 class="title">Pascal Allen<hr></h2>

@stop

@section('bottom-script')

@stop
