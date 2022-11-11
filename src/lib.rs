extern crate alloc;
use ::value::{Secrets, Value};
use lazy_static::lazy_static;
use regex::Regex;
use std::borrow::Borrow;
use std::cell::RefCell;
use std::collections::BTreeMap;
use std::ffi::CStr;
use std::ffi::CString;
use vrl::diagnostic::Formatter;
use vrl::Program;
use vrl::TimeZone;
use vrl::{state, Runtime, TargetValueRef};

use alloc::vec::Vec;
use std::mem::MaybeUninit;
use std::slice;

lazy_static! {
    static ref RE: Regex = Regex::new(r"\\w{4}\\s\\w{3}\\s\\w").unwrap();
    static ref VRL_PROGRAM: Program = compile_vrl_static();
}

thread_local! {static RUNTIME: RefCell<Runtime> = RefCell::new(Runtime::new(state::Runtime::default()));}

pub fn compile_vrl_static() -> Program {
    // let program = r#"."#;
    let program = r#". = replace(string!(.), r'\b\w{4}\b', "rust", 1)"#;
    let functions = vrl_stdlib::all();
    match vrl::compile(&program, &functions) {
        Ok(res) => res.program,
        Err(err) => {
            let f = Formatter::new(&"", err);
            panic!("{:#}", f)
        }
    }
}

#[no_mangle]
pub extern "C" fn compile_vrl(input: *const libc::c_char) -> *mut Program {
    let program_string = unsafe { CStr::from_ptr(input) }.to_str().unwrap();
    let functions = vrl_stdlib::all();
    match vrl::compile(&program_string, &functions) {
        Ok(res) => {
            return Box::into_raw(Box::new(res.program));
        },
        Err(err) => {
            let f = Formatter::new(&"", err);
            panic!("{:#}", f)
        }
    }
}


pub fn run_vrl_static(s: &str) -> String {
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

pub fn run_vrl(s: &str, program: &Program) -> String {
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
            .resolve(&mut target, program, &TimeZone::Local);
    });

    return output.unwrap().to_string();
}

#[no_mangle]
pub extern "C" fn transform(input: *const libc::c_char) -> *const libc::c_char {
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let replaced = RE.replacen(inpt.to_str().unwrap(), 1, "rust");
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
pub extern "C" fn transform_vrl(input: *const libc::c_char, program: *mut Program) -> *const libc::c_char {
    let prog =  unsafe { program.as_ref().unwrap() };
    let inpt: &CStr = unsafe { CStr::from_ptr(input) };
    let output = run_vrl(inpt.to_str().unwrap(), prog);
    let c_str = CString::new(output.as_bytes()).expect("CString::new failed");
    return c_str.into_raw();
}

// Wasm Integration Below
//
/// WebAssembly export that accepts a string (linear memory offset, byteCount)
/// and creates a copy, then writes that copy back into the same place in
/// memory. It returns the length of the string that was just written.
///
#[cfg_attr(all(target_arch = "wasm32"), export_name = "regex_wasm")]
#[no_mangle]
pub unsafe extern "C" fn _regex_wasm(ptr: u32, len: u32) -> u32 {
    let name = &ptr_to_string(ptr, len);

    let output = RE.replacen(name, 1, "rust");
    store_string_at_ptr(&output, ptr);

    output.len() as u32
}
/// WebAssembly export that accepts a string (linear memory offset, byteCount)
/// and creates a copy, then writes that copy back into the same place in
/// memory. It returns the length of the string that was just written.
///
#[cfg_attr(all(target_arch = "wasm32"), export_name = "vrl_wasm")]
#[no_mangle]
pub unsafe extern "C" fn _vrl_wasm_buffered(ptr: u32, len: u32) -> u32 {
    let name = &ptr_to_string(ptr, len);

    let output = run_vrl_static(name);
    store_string_at_ptr(&output, ptr);

    output.len() as u32
}

/// WebAssembly export that accepts a string (linear memory offset, byteCount)
/// and creates a copy, then writes that copy back into the same place in
/// memory. It returns the length of the string that was just written.
///
#[cfg_attr(all(target_arch = "wasm32"), export_name = "noop_wasm")]
#[no_mangle]
pub unsafe extern "C" fn _noop_wasm_buffered(ptr: u32, len: u32) -> u32 {
    let name = &ptr_to_string(ptr, len);
    let new_string = String::from(name); // the no-op
    store_string_at_ptr(&new_string, ptr);

    new_string.len() as u32
}
/// WebAssembly export that accepts a string (linear memory offset, byteCount)
/// and returns a pointer/size pair packed into a u64.
///
/// Note: The return value is leaked to the caller, so it must call
/// [`deallocate`] when finished.
/// Note: This uses a u64 instead of two result values for compatibility with
/// WebAssembly 1.0.
#[cfg_attr(
    all(target_arch = "wasm32"),
    export_name = "noop_wasm_dynamic_allocation"
)]
#[no_mangle]
pub unsafe extern "C" fn _noop_wasm(ptr: u32, len: u32) -> u64 {
    let name = &ptr_to_string(ptr, len);
    let new_string = String::from(name); // the no-op
    let (ptr, len) = string_to_ptr(&new_string);
    // Note: This changes ownership of the pointer to the external caller. If
    // we didn't call forget, the caller would read back a corrupt value. Since
    // we call forget, the caller must deallocate externally to prevent leaks.
    std::mem::forget(new_string);
    return ((ptr as u64) << 32) | len as u64;
}

// WASM String-related helper functions
/// Returns a string from WebAssembly compatible numeric types representing
/// its pointer and length.
unsafe fn ptr_to_string(ptr: u32, len: u32) -> String {
    let slice = slice::from_raw_parts_mut(ptr as *mut u8, len as usize);
    let utf8 = std::str::from_utf8_unchecked_mut(slice);
    return String::from(utf8);
}

/// Stores the given string 's' at the memory location pointed to by 'ptr'
/// This assumes no buffer overflows - here be dragons.
unsafe fn store_string_at_ptr(s: &str, ptr: u32) {
    // Create a mutable slice of u8 pointing at the buffer given as 'ptr'
    // with a length of the string we're about to copy into it
    let dest = slice::from_raw_parts_mut(ptr as *mut u8, s.len() as usize);
    dest.copy_from_slice(s.as_bytes());
}

/// Returns a pointer and size pair for the given string in a way compatible
/// with WebAssembly numeric types.
///
/// Note: This doesn't change the ownership of the String. To intentionally
/// leak it, use [`std::mem::forget`] on the input after calling this.
unsafe fn string_to_ptr(s: &String) -> (u32, u32) {
    return (s.as_ptr() as u32, s.len() as u32);
}

// WASM Memory-related helper functinos
//
// TODO explore using lol_alloc instead of default rust allocator
/// WebAssembly export that allocates a pointer (linear memory offset) that can
/// be used for a string.
///
/// This is an ownership transfer, which means the caller must call
/// [`deallocate`] when finished.
#[cfg_attr(all(target_arch = "wasm32"), export_name = "allocate")]
#[no_mangle]
pub extern "C" fn _allocate(size: u32) -> *mut u8 {
    allocate(size as usize)
}

/// Allocates size bytes and leaks the pointer where they start.
fn allocate(size: usize) -> *mut u8 {
    // Allocate the amount of bytes needed.
    let vec: Vec<MaybeUninit<u8>> = Vec::with_capacity(size);

    // into_raw leaks the memory to the caller.
    Box::into_raw(vec.into_boxed_slice()) as *mut u8
}

/// WebAssembly export that deallocates a pointer of the given size (linear
/// memory offset, byteCount) allocated by [`allocate`].
#[cfg_attr(all(target_arch = "wasm32"), export_name = "deallocate")]
#[no_mangle]
pub unsafe extern "C" fn _deallocate(ptr: u32, size: u32) {
    deallocate(ptr as *mut u8, size as usize);
}

/// Retakes the pointer which allows its memory to be freed.
unsafe fn deallocate(ptr: *mut u8, size: usize) {
    // TODO - should this be Box::from_raw? (see Box::into_raw docs)
    let _ = Vec::from_raw_parts(ptr, 0, size);
}
