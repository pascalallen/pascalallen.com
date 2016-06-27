<?php

class NationalParksController extends \BaseController {

	/**
	 * Display a listing of nationalparks
	 *
	 * @return Response
	 */
	public function index()
	{
		$national_parks = NationalPark::paginate(2);

		return View::make('national_parks.index')->with(['national_parks' => $national_parks]);
	}

	/**
	 * Show the form for creating a new nationalpark
	 *
	 * @return Response
	 */
	public function create()
	{
		return View::make('national_parks.create');
	}

	/**
	 * Store a newly created nationalpark in storage.
	 *
	 * @return Response
	 */
	public function store()
	{
		$national_park = new NationalPark();
		return $this->validateAndSave($national_park);
	}

	protected function validateAndSave($national_park)
	{
		$validator = Validator::make(Input::all(), NationalPark::$rules);

		if ($validator->fails()) {
	        // validation failed, redirect to the tutorial create page with validation errors and old inputs
	        return Redirect::back()->withInput()->withErrors($validator);
	    } else {
	    	
			$national_park->name = Input::get('name');
			$national_park->location = Input::get('location');
			$national_park->date_established = Input::get('date_established');
			$national_park->area_in_acres = Input::get('area_in_acres');
			$national_park->description = Input::get('description');

			$result = $national_park->save();

			if($result) {
				Session::flash('successMessage', 'Your National Park has been saved.');
				return Redirect::action('NationalParksController@index');
			} else {
				return Redirect::back()->withInput();
			}
		}
	}

	/**
	 * Display the specified nationalpark.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function show($id)
	{
		$national_park = NationalPark::findOrFail($id);

		return View::make('national_parks.show', compact('national_park'));
	}

	/**
	 * Show the form for editing the specified nationalpark.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function edit($id)
	{
		$national_park = NationalPark::find($id);

		return View::make('national_parks.edit', compact('national_park'));
	}

	/**
	 * Update the specified nationalpark in storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function update($id)
	{
		$national_park = NationalPark::findOrFail($id);

		$validator = Validator::make($data = Input::all(), NationalPark::$rules);

		if ($validator->fails())
		{
			return Redirect::back()->withErrors($validator)->withInput();
		}

		$national_park->update($data);

		return Redirect::route('national_parks.index');
	}

	/**
	 * Remove the specified nationalpark from storage.
	 *
	 * @param  int  $id
	 * @return Response
	 */
	public function destroy($id)
	{
		NationalPark::destroy($id);

		return Redirect::route('national_parks.index');
	}

}
