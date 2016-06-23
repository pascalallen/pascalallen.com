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
		    margin-left: 73%;
		    top: 8%;
		}

		#two{
		    width: 5%;
		    margin-left: 73%;
		    top: 18%;
		}

		#three{
		    width: 5%;
		    margin-left: 61%;
		    top: 8%;
		}

		#four{
		    width: 5%;
		    margin-left: 61%;
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
				<a href="national_parks.php"><img src="img/folder.png" class="folder" id="two"></a>
				<a href=""><img src="img/folder.png" class="folder" id="three"></a>
				<a href=""><img src="img/folder.png" class="folder" id="four"></a>
			</div>
		</div>
	</div>
	
@stop

@section('bottom-script')

	<script type="text/javascript">

	</script>

@stop
