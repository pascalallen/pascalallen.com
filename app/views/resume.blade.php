@extends('layouts.master')

@section('top-script')
	<link rel="stylesheet" type="text/css" href="resume.css">
@stop

@section('content')
<div class="site-wrapper">

	<div class="site-wrapper-inner">

		<div class="cover-container">

			<div class="masthead clearfix">
				<div class="inner">
					<h3 class="masthead-brand">Pascal Allen</h3>
					<nav>
						<ul class="nav masthead-nav">
							<li class="active"><a href="#">Home</a></li>
							<li><a href="#">Features</a></li>
							<li><a href="#">Contact</a></li>
						</ul>
					</nav>
				</div>
			</div>

			<div class="inner cover">
				<h1 class="cover-heading">Welcome To My Resume</h1>
				<p class="lead">Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod
				tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,
				quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
				consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse
				cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non
				proident, sunt in culpa qui officia deserunt mollit anim id est laborum.</p>
				<p class="lead">
					<a href="#" class="btn btn-lg btn-default">Learn more</a>
				</p>
			</div>

			<div class="mastfoot">
				<div class="inner">
					<p>Designed by Pascal Allen</p>
				</div>
			</div>

		</div>

	</div>

</div>
@stop