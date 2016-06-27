<?php

class NationalParksController extends \BaseController {

	/**
	 * Display a listing of nationalparks
	 *
	 * @return Response
	 */
	public function index()
	{
		$nationalparks = Nationalpark::all();

		return View::make('nationalparks.index', compact('nationalparks'));
	}

	/**
	 * Show the form for creating a new nationalpark
	 *
	 * @return Response
	 */
	public function create()
	{
		return View::make('nationalparks.create');
	}

	/**
	 * Store a newly created nationalpark in storage.
	 *
	 * @return Response
	 */
	public function store()
	{
		$validator = Validator::make($data = Input::all(), Nationalpark::$rules);

		if ($validator->fails())
		{
			return Redirect::back()->withErrors($validator)->withInput();
		}

		Nationalpark::create($data);

		return Redirect::route('nationalparks.index');
	}

	/**
	 * Display the specified nationalpark.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function show($id)
	{
		$nationalpark = Nationalpark::findOrFail($id);

		return View::make('nationalparks.show', compact('nationalpark'));
	}

	/**
	 * Show the form for editing the specified nationalpark.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function edit($id)
	{
		$nationalpark = Nationalpark::find($id);

		return View::make('nationalparks.edit', compact('nationalpark'));
	}

	/**
	 * Update the specified nationalpark in storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function update($id)
	{
		$nationalpark = Nationalpark::findOrFail($id);

		$validator = Validator::make($data = Input::all(), Nationalpark::$rules);

		if ($validator->fails())
		{
			return Redirect::back()->withErrors($validator)->withInput();
		}

		$nationalpark->update($data);

		return Redirect::route('nationalparks.index');
	}

	/**
	 * Remove the specified nationalpark from storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function destroy($id)
	{
		Nationalpark::destroy($id);

		return Redirect::route('nationalparks.index');
	}

}
