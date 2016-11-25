<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;

use App\Http\Requests;
use App\User;
use Log;
use Input;
use Validator;
use Session;
use Redirect;
use Auth;

class UsersController extends Controller
{
 //    public function __construct()
	// {
	// 	parent::__construct();
	// 	$this->beforeFilter('auth', array('except' => array('create', 'store')));
	// }
	/**
	 * Display a listing of the resource.
	 *
	 * @return Response
	 */
	public function index()
	{
		$user = Auth::user();
		return view('users.index')->with('user', $user);
	}
	/**
	 * Show the form for creating a new resource.
	 *
	 * @return Response
	 */
	public function create()
	{
		return view('users.create');
	}
	/**
	 * Store a newly created resource in storage.
	 *
	 * @return Response
	 */
	public function store()
	{
		$user = new User();
		$validator = Validator::make(Input::all(), User::$rules);
		if ($validator->fails()) {
	        // validation failed, redirect to the tutorial create page with validation errors and old inputs
	        Session::flash('errorMessage', 'Validation failed');
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
			$user->name = Input::get('name');
			$user->password = Input::get('password');
			$user->email = Input::get('email');
			$user->admin = 1;
			$user->save();
			Auth::login($user);
			$user = Auth::user();
			Session::flash('successMessage', 'Come on in!');
			return Redirect::action('UsersController@show', $user->id);
		}
	}
	/**
	 * Display the specified resource.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function show($id)
	{
		$user = User::find($id);
		
		return view('users.show', compact('user'));
	}
	/**
	 * Show the form for editing the specified resource.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function edit($id)
	{
		if(Auth::id() != $id)
		{
			Log::info('This user id is not equal');
			return Redirect::action('UsersController@show', Auth::id());
		}
		$post = Post::with('user')->get();
		$user = User::find($id);
		return view('users.edit', compact('user'));
	}
	/**
	 * Update the specified resource in storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function update($id)
	{
		$user = User::find($id);
		Log::info(Input::all());
		$validator = Validator::make(Input::all(), User::$rules);
		if ($validator->fails()) {
	        // validation failed, redirect to the tutorial create page with validation errors and old inputs
	        Session::flash('errorMessage', 'Validation failed');
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
			$user->name = Input::get('name');
			$user->password = Input::get('password');
			$user->email = Input::get('email');
			$user->save();
			Auth::login($user);
			$user = Auth::user();
			Session::flash('successMessage', 'Your user has been saved.');
			return Redirect::action('UsersController@show', $user->id);
		}
	}
	/**
	 * Remove the specified resource from storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function destroy($id)
	{
		$user = User::find($id);
		$user->delete();
		return Redirect::action('HomeController@showWelcome');
	}
}
