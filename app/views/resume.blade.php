@extends('layouts.master')

@section('top-script')
	{{-- CUSTOM CSS --}}
	<link rel="stylesheet" type="text/css" href="/css/resume.css">
	{{-- CUSTOM FONT --}}
	<link href='https://fonts.googleapis.com/css?family=Ubuntu' rel='stylesheet' type='text/css'>
	{{-- FONT AWESOME --}}
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/font-awesome/4.5.0/css/font-awesome.min.css">
@stop

@section('content')
<div class="container">
	<div class="subcontainer">
		<a href="#">
			<div class="portfolio">
				<div class="port-text">Visit My Porfolio<br><i class="fa fa-folder"></i></div>
			</div>
		</a>
		<div class="img-container">
			<img class="image-circle" src="../img/resume_pic.jpg">
		</div>
		<div class="header">Pascal Allen</div>
		<div class="subheader">Full-Stack Web Developer</div>
		<hr>
		<div class="nav">
			<a class="btn btn-lg btn-success" href="#"><i class="fa fa-envelope-o fa-lg"></i></a>
			<a class="btn btn-lg btn-success" href="#"><i class="fa fa-linkedin fa-lg"></i></a>
		</div>
		<div class="resume">
			<img class="resume-pdf" src="../img/web_dev_resume.jpg">
		</div>
		<br>
		<div class="footer-icon">
			<a class="btn btn-lg btn-success" href="#"><i class="fa fa-file-pdf-o fa-lg"></i></a>
			<a class="btn btn-lg btn-success" href="#"><i class="fa fa-file-word-o fa-lg"></i></a>
			<a class="btn btn-lg btn-success" href="#"><i class="fa fa-file-text-o fa-lg"></i></a>
		</div>
		<hr>
		<footer>Designed by Pascal Allen. <a href="https://github.com/pascalallen"><i class="fa fa-github-alt fa-lg"></i></a></footer>
		<br>
	</div>
</div>
@stop