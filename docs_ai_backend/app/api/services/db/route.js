import { NextResponse } from "next/server";
import { populateDB } from "../../../../src/utils/populateDB"

export async function POST(req, res) {

  try {


    await populateDB();


    return NextResponse.json({ data: 'Success', status: 200 });

  } catch (error) {
    console.log('!');
    console.log(error);
    return NextResponse.json({ error: error }, { status: 500 })
  }

}
