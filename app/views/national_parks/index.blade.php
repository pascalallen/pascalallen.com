@extends('layouts.master')

@section('top-script')
	
	<!-- CUSTOM FONT -->
	<link href='https://fonts.googleapis.com/css?family=Rock+Salt' rel='stylesheet' type='text/css'>

	<!-- TITLE IMG -->
	<link rel="shortcut icon" href="img/codeup-arrow.png">

	<!-- CUSTOM CSS -->
	<link rel="stylesheet" type="text/css" href="css/national_parks.css">

@stop

@section('content')

	<div class="site-wrapper">
		<div class="container">
			<h2>National Parks</h2>
			<div class="row">
				<div class="col-xs-8 col-xs-offset-2">
					<table class="table">
						<thead>
							<tr>
							<th>Name</th>
							<th>Location</th>
							<th>Established</th>
							<th>Acres</th>
							<th>Description</th>
							<th></th>
							</tr>
					    </thead>
						<?php foreach ($national_parks as $park) : ?>
							<tbody>
								<tr class="table-bordered">
								    <td><?= $park['name'] ?> </td>
								    <td><?= $park['location'] ?> </td>
								    <td><?= $park['date_established'] ?> </td>
								    <td><?= $park['area_in_acres'] ?> </td>
								    <td><?= $park['description'] ?> </td>
								    <td>
							    		<form role="form" method="POST">
											<button type="submit" class="btn btn-info btn-xs" value="<?= $park['id'] ?>" name="id">Delete</button>
										</form>
									</td>
								</tr>
							</tbody>
						<?php endforeach; ?>
					</table>
					{{ $national_parks->links() }}
				</div>
			</div>


			<div class="row">
				<h3 class="col-xs-4 col-xs-offset-4">Submit a park:</h3>
				<div class="col-xs-4 col-xs-offset-4">
					{{ Form::open(array('method' => 'post', 'action' => 'NationalParksController@store')) }}

						<div class="form-group {{ ($errors->has('name')) ? 'has-error' : '' }}">
							{{ $errors->first('name', '<div class="alert alert-danger">:message</div>') }}
							{{ Form::label('name', 'Name') }}
							{{ Form::text('name', null, ['class' => 'form-control', 'placeholder' => 'Name']) }}
						</div>

						<div class="form-group {{ ($errors->has('location')) ? 'has-error' : '' }}">
							{{ $errors->first('location', '<div class="alert alert-danger">:message</div>') }}
							{{ Form::label('location', 'Location') }}
							{{ Form::text('location', null, ['class' => 'form-control', 'placeholder' => 'Location']) }}
						</div>

						<div class="form-group {{ ($errors->has('date_established')) ? 'has-error' : '' }}">
							{{ $errors->first('date_established', '<div class="alert alert-danger">:message</div>') }}
							{{ Form::label('date_established', 'Date Established') }}
							{{ Form::text('date_established', null, ['class' => 'form-control', 'placeholder' => 'Date Established']) }}
						</div>

						<div class="form-group {{ ($errors->has('area_in_acres')) ? 'has-error' : '' }}">
							{{ $errors->first('area_in_acres', '<div class="alert alert-danger">:message</div>') }}
							{{ Form::label('area_in_acres', 'Area In Acres') }}
							{{ Form::text('area_in_acres', null, ['class' => 'form-control', 'placeholder' => 'Area In Acres']) }}
						</div>

						<div class="form-group {{ ($errors->has('description')) ? 'has-error' : '' }}">
							{{ $errors->first('description', '<div class="alert alert-danger">:message</div>') }}
							{{ Form::label('description', 'Description') }}
							{{ Form::textarea('description', null, ['class' => 'form-control', 'placeholder' => 'Description']) }}
						</div>

						<button type="submit" class="btn btn-default">Submit</button>

					{{ Form::close() }}
				</div>
			</div>
		</div>
	</div>

@stop

@section('bottom-script')

	<!-- CUSTOM JAVASCRIPT -->
	<script src="js/national_parks.js"></script>

@stop
