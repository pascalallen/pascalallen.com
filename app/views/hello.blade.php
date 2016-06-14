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
		    z-index: -1;
		}

	</style>

@stop

@section('content')

	<div class="container">
  		<div class="row">
    		<div class="span12">
				<img src="img/transparent-macbook.png" class="macbook">
				<img src="img/folder.png" class="folder">
			</div>
		</div>
	</div>
	
@stop

@section('bottom-script')

	<script type="text/javascript">

	</script>

@stop
