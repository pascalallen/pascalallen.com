@extends('layouts.master')

@section('top-script')
	<link rel="stylesheet" type="text/css" href="/css/resume.css">
	<link href='https://fonts.googleapis.com/css?family=Ubuntu' rel='stylesheet' type='text/css'>
@stop

@section('content')
<div class="container">
	<div class="img-container">
		<img class="image-circle" src="../img/resume_pic.jpg">
	</div>
	<div class="header">Pascal Allen</div>
	<div class="subheader">Full-Stack Web Developer</div>
	<hr>
	<div class="resume">
		<img class="resume-pdf" src="../img/web_dev_resume.jpg">
	</div>
	{{-- put buttons to download pdf and word document type resume files here --}}
</div>
@stop