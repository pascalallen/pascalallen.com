@extends('layouts.master')

@section('top-script')

    <style type="text/css">
        .row {
            margin: 0;
            text-align: center;
        }
        .login {
            position: absolute;
            margin-top: 150px;
        }
    </style>

@stop

@section('content')

    <div class="row">
        <div class="col-xs-4 col-xs-offset-4 login">
            {!! Form::open(array('action' => 'HomeController@postLogin')) !!}
                <div class="{!! ($errors->has('email')) ? 'has-error' : '' !!} form-group">
                    {!! $errors->first('email', '<div class="alert alert-danger">:message</div>') !!}
                    {!! Form::label('email', 'Email') !!}
                    {!! Form::email('email', Input::old('email'), ['class' => 'form-control', 'placeholder'=>'Email']) !!}
                </div>
                <div class="{!! ($errors->has('password')) ? 'has-error' : '' !!} form-group">
                    {!! $errors->first('password', '<div class="alert alert-danger">:message</div>') !!}
                    {!! Form::label('password', 'Password') !!}
                    {!! Form::password('password', ['class' => 'form-control', 'placeholder'=>'Password']) !!}
                </div>
                <button type="submit" class="btn btn-default">Submit</button>
            {!! Form::close() !!}
            <a href="{!! action('UsersController@create') !!}">Not a member? Sign up!</a>
        </div>
    </div>

@stop

@section('bottom-script')

@stop