@extends('layouts.master')

@section('top-script')

	<style>

		.macbook{
			width: 100%;
		}

		.folder{
			position: absolute;
		    top: 10%;
		    right: 18%;
		    width: 5%;
		}


	</style>

@stop

@section('content')

	<div class="container">
  		<div class="row">
    		<div class="span12">
				<img src="img/transparent-macbook.png" class="macbook">
				<a href="{{ action('HomeController@tattooArtistProject') }}"><img src="img/folder.png" class="folder"></a>
			</div>
		</div>
	</div>
	
@stop

@section('bottom-script')

	<script type="text/javascript">

	</script>

@stop
