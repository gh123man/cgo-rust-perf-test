use ::value::{Secrets, Value};
use lazy_static::lazy_static;
use regex::Regex;
use std::cell::RefCell;
use std::collections::BTreeMap;
use std::ffi::CStr;
use std::ffi::CString;
use vrl::Program;
use vrl::TimeZone;
use vrl::{state, Runtime, TargetValueRef};

lazy_static! {
    static ref RE: Regex = Regex::new(r"\b\w{4}\b").unwrap();
    static ref VRL_PROGRAM: Program = compile_vrl();
}

thread_local! {static RUNTIME: RefCell<Runtime> = RefCell::new(Runtime::new(state::Runtime::default()));}

pub fn compile_vrl() -> Program {
    // let program = r#"."#;
    let program = r#". = replace(string!(.), r'\b\w{4}\b', "rust", 1)"#;
    return vrl::compile(&program, &vrl_stdlib::all()).unwrap().program;
}

pub fn run_vrl(s: &str) -> String {
    let mut value: Value = Value::from(s);
    let mut metadata = Value::Object(BTreeMap::new());
    let mut secrets = Secrets::new();
    let mut target = TargetValueRef {
        value: &mut value,
        metadata: &mut metadata,
        secrets: &mut secrets,
    };

    let output = RUNTIME.with(|r| {
        // r.borrow_mut().clear();
        return r
            .borrow_mut()
            .resolve(&mut target, &VRL_PROGRAM, &TimeZone::Local);
    });

    return output.unwrap().to_string();
}

#[no_mangle]
pub extern "C" fn transform(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let replaced = RE.replace(inpt.to_str().unwrap(), "rust");
    let c_str = CString::new(replaced.as_bytes()).expect("CString::new failed");
    return c_str.into_raw();
}

#[no_mangle]
pub extern "C" fn noop(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let c_str = CString::new(inpt.to_str().unwrap()).expect("CString::new failed");
    return c_str.into_raw();
}

#[no_mangle]
pub extern "C" fn transform_vrl(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let output = run_vrl(inpt.to_str().unwrap());
    let c_str = CString::new(output.as_bytes()).expect("CString::new failed");
    return c_str.into_raw();
}

#[no_mangle]
pub extern "C" fn passthrough(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let rust_str = inpt.to_str().unwrap().to_string();
    let c_str = CString::new(rust_str).expect("CString::new failed");

    return c_str.into_raw();
}
