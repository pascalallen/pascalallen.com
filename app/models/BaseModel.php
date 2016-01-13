<?php

use Carbon\Carbon;

class BaseModel extends Eloquent {

	public function getUpdatedAtAttribute($date)
	{
	    return Carbon::createFromFormat('Y-m-d H:i:s', $date)->diffForHumans();
	}

	public function getCreatedAtAttribute($value)
	{
	    $utc = Carbon::createFromFormat($this->getDateFormat(), $value);
	    return $utc->setTimezone('America/Chicago');
	}

}