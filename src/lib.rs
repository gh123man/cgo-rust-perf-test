use std::ffi::CStr;
use std::ffi::CString;
use regex::Regex;
use lazy_static::lazy_static;

#[no_mangle]
pub extern "C" fn callme() {
    println!("Hello, world!");
}


lazy_static! {
    static ref RE: Regex = Regex::new(r"\b\w{4}\b").unwrap();
}

#[no_mangle]
pub extern "C" fn transform(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };

    // let mut rust_str = inpt.to_str().unwrap().to_string();

    // let re = Regex::new(r"\b\w{4}\b").unwrap();
    let replaced = RE.replace(inpt.to_str().unwrap(), "rust");

    let c_str = CString::new(replaced.as_bytes()).expect("CString::new failed");
    // let c_str = CString::new(rust_str).expect("CString::new failed");
    return c_str.into_raw();
}
