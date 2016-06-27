@extends('layouts.master')

@section('top-script')

	<style>

		.macbook{
			width: 100%;
		}

		.folder{
			position: absolute;
		}

		#one{
		    width: 5%;
		    left: 78%;
		    top: 8%;
		}

		#two{
		    width: 5%;
		    left: 78%;
		    top: 18%;
		}

		#three{
		    width: 5%;
		    left: 70%;
		    top: 8%;
		}

		#four{
		    width: 5%;
		    left: 70%;
		    top: 18%;
		}

		#five{
		    width: 5%;
		    left: 62%;
		    top: 8%;
		}

		#six{
		    width: 5%;
		    left: 62%;
		    top: 18%;
		}

		#seven{
		    width: 5%;
		    left: 54%;
		    top: 8%;
		}

		#eight{
		    width: 5%;
		    left: 54%;
		    top: 18%;
		}

	</style>

@stop

@section('content')

	<div class="container">
  		<div class="row">
    		<div class="span12">
				<img src="img/transparent-macbook.png" class="macbook">
				<a href="{{ action('HomeController@tattooArtistProject') }}"><img src="img/folder.png" class="folder" id="one"></a>
				<a href="{{ action('NationalParksController@index') }}"><img src="img/folder.png" class="folder" id="two"></a>
				<a href="{{ action('HomeController@coleman') }}"><img src="img/folder.png" class="folder" id="three"></a>
				<a href="{{ action('HomeController@showResume') }}"><img src="img/folder.png" class="folder" id="four"></a>
				<a href=""><img src="img/folder.png" class="folder" id="five"></a>
				<a href=""><img src="img/folder.png" class="folder" id="six"></a>
				<a href=""><img src="img/folder.png" class="folder" id="seven"></a>
				<a href=""><img src="img/folder.png" class="folder" id="eight"></a>
			</div>
		</div>
	</div>
	
@stop

@section('bottom-script')

	<script type="text/javascript">

	</script>

@stop
