<!DOCTYPE html>
<html>
	<head>	
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
	    <meta name="viewport" content="width=device-width, initial-scale=1">
	    <title>User/Admin Portal</title>
		{{-- BOOTSTRAP CSS --}}
		<link rel="stylesheet" href="/css/bootstrap.min.css">
		{{-- FONT AWESOME --}}
		<link rel="stylesheet" href="/css/font-awesome.min.css">
		{{-- CUSTOM CSS --}}
		<link rel="stylesheet" type="text/css" href="/css/main.css">
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
		<script src="/js/jquery-3.1.1.min.js"></script>
		<!-- BOOTSTRAP JS -->
		<script src="/js/bootstrap.min.js"></script>
		{{-- AngularJS --}}
		<script src="/js/angular.min.js"></script>
		<!-- CUSTOM JS BELOW -->
		<script src="/js/main.js"></script>
		@yield('bottom-script')
	</body>
</html>