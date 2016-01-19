<!DOCTYPE html>
<html>
	<head>	
		<title></title>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
	    <meta name="viewport" content="width=device-width, initial-scale=1">
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
		<nav id="mainNav" class="navbar navbar-default navbar-fixed-top affix-top">
	        <div class="container-fluid">
	            <!-- Brand and toggle get grouped for better mobile display -->
	            <div class="navbar-header">
	                <a class="navbar-brand page-scroll" href="/posts">Bloggy Blog</a>
	            </div>
				<div class="col-sm-3 col-md-3 pull-left">
					@if (Auth::check() && Request::is('posts'))

						{{ Form::open(array('action' => array('PostController@index'), 'method' => 'GET')) }}
							<div class="input-group">
								{{ Form::text('search', $search, ['class' => 'form-control', 'placeholder' => 'Search']) }}
								<div class="input-group-btn">
									<button class="btn btn-default" type="submit"><i class="fa fa-search"></i></button>
								</div>
							</div>

						{{ Form::close() }}

					@endif
				</div>
	            <!-- Collect the nav links, forms, and other content for toggling -->
	            <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
	                <ul class="nav navbar-nav navbar-right">
	                    <li class="">
	                    	<a class="page-scroll" href="/posts/create">Create A Post</a>
	                    </li>
	                    <li class="">
	                        <a class="page-scroll" href="/portfolio">Portfolio</a>
	                    </li>
	                    <li class="">
	                        <a class="page-scroll" href="/resume">Resume</a>
	                    </li>
	                    @if (Auth::check())

		                    <li class="">
		                        <a class="page-scroll" href="{{ action('HomeController@getLogout') }}">Logout</a>
		                    </li>

	                    @else

		                    <li class="">
		                        <a class="page-scroll" href="/login">Login</a>
		                    </li>
		                    <li class="">
		                        <a class="page-scroll" href="/login">Sign Up</a>
		                    </li>

	                    @endif
	                </ul>
	            </div>
	            <!-- /.navbar-collapse -->
	        </div>
	        <!-- /.container-fluid -->
	    </nav>
		<div class="message">
		    @if (Session::has('successMessage'))
			    <div class="alert alert-success">{{{ Session::get('successMessage') }}}</div>
			@endif
			@if (Session::has('errorMessage'))
			    <div class="alert alert-danger">{{{ Session::get('errorMessage') }}}</div>
			@endif
		</div>
		@yield('content')
		<!-- JQUERY -->
		<script src="/js/jquery-2.1.4.min.js"></script>
		<!-- BOOTSTRAP JS -->
		<script src="/js/bootstrap.min.js"></script>
		<!-- CUSTOM JS -->
		{{-- <script src="js/main.js"></script> --}}

		@yield('bottom-script')
	</body>
</html>
