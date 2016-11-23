<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
	    <meta name="viewport" content="width=device-width, initial-scale=1">
	    <title>Pascal Allen</title>

	    {{-- TITLE ICON --}}
		<link rel="shortcut icon" type="image/x-icon" href="img/apple.ico" />
		
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

		{{-- @include('partials.navbar') --}}
        
		@if (Session::has('successMessage'))
		    <div class="alert alert-success col-xs-4 col-xs-offset-4">{{{ Session::get('successMessage') }}}</div>
		@endif
		
		@if (Session::has('errorMessage'))
		    <div class="alert alert-danger col-xs-4 col-xs-offset-4">{{{ Session::get('errorMessage') }}}</div>
		@endif

		@yield('content')

		{{-- @include('partials.footer') --}}

		<!-- ANGULARJS -->
		<script src="/js/angular.min.js"></script>

		<!-- JQUERY -->
		<script src="/js/jquery-2.1.4.min.js"></script>

		<!-- BOOTSTRAP JS -->
		<script src="/js/bootstrap.min.js"></script>

		<!-- CUSTOM JS BELOW -->
		<script src="/js/main.js"></script>

		@yield('bottom-script')

	</body>