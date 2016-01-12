<!DOCTYPE html>
<html>
	<head>	
		<title></title>
		<meta charset="UTF-8">
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
	    <meta name="viewport" content="width=device-width, initial-scale=1">
		<!-- BOOTSTRAP CSS -->
		<link href="/css/bootstrap.min.css" rel="stylesheet">
		<!-- CUSTOM CSS -->
		<link rel="stylesheet" type="text/css" href="/css/main.css">
		
		@yield('top-script')
	</head>
	<body>
		<nav id="mainNav" class="navbar navbar-default navbar-fixed-top affix-top">
	        <div class="container-fluid">
	            <!-- Brand and toggle get grouped for better mobile display -->
	            <div class="navbar-header">
	                <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1">
	                    <span class="sr-only">Toggle navigation</span>
	                    <span class="icon-bar"></span>
	                    <span class="icon-bar"></span>
	                    <span class="icon-bar"></span>
	                </button>
	                <a class="navbar-brand page-scroll" href="/posts">Bloggy Blog</a>
	            </div>

	            <!-- Collect the nav links, forms, and other content for toggling -->
	            <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">
	                <ul class="nav navbar-nav navbar-right">
	                    <li class="">
	                    	<a class="page-scroll" href="/posts/create">Create A Post</a>
	                    </li>
	                    <li class="">
	                        <a class="page-scroll" href="/posts">Home</a>
	                    </li>
	                    <li class="">
	                        <a class="page-scroll" href="/portfolio">Portfolio</a>
	                    </li>
	                    <li class="">
	                        <a class="page-scroll" href="/resume">Resume</a>
	                    </li>
	                </ul>
	            </div>
	            <!-- /.navbar-collapse -->
	        </div>
	        <!-- /.container-fluid -->
	    </nav>
		@yield('content')
		<!-- JQUERY -->
		<script src="/js/jquery-2.1.4.min.js"></script>
		<!-- BOOTSTRAP JS -->
		<script src="/js/bootstrap.min.js"></script>
		<!-- CUSTOM JS -->
		<script src="js/main.js"></script>

		@yield('bottom-script')
	</body>
</html>
