@extends('layouts.master')

@section('top-script')

@stop

@section('content')
	<div data-ng-app="sampleApp">
		{{-- placeholder for view --}}
		<div data-ng-view=""></div>
	</div>

@stop

@section('bottom-script')

	<script>

	</script>

@stop
