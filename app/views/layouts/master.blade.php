<!DOCTYPE html>
<html>
	<head>	
		<title>Dirt Swaps</title>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
	    <meta name="viewport" content="width=device-width, initial-scale=1">
	    <title>Dirt Swaps</title>

		{{-- BOOTSTRAP CSS --}}
		<link href="/css/bootstrap.min.css" rel="stylesheet">

		{{-- CUSTOM CSS --}}
		<link rel="stylesheet" type="text/css" href="/css/main.css">

		{{-- CUSTOM FONT --}}
		<link href='https://fonts.googleapis.com/css?family=Ubuntu' rel='stylesheet' type='text/css'>

		{{-- FONT AWESOME --}}
		<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/4.5.0/css/font-awesome.min.css">
		
		@yield('top-script')
	</head>
	<body>
		@include('partials.navbar')

		@yield('content')

		@if (Session::has('successMessage'))
		    <div class="alert alert-success">{{{ Session::get('successMessage') }}}</div>
		@endif
		
		@if (Session::has('errorMessage'))
		    <div class="alert alert-danger">{{{ Session::get('errorMessage') }}}</div>
		@endif

		@include('partials.footer')

		<!-- JQUERY -->
		<script src="/js/jquery-2.1.4.min.js"></script>

		<!-- BOOTSTRAP JS -->
		<script src="/js/bootstrap.min.js"></script>

		<!-- CUSTOM JS BELOW -->
		<script src="/js/main.js"></script>

		@yield('bottom-script')

	</body>