use anyhow::Result;
use std::{
    collections::HashMap,
    ffi::OsString,
    fs,
    path::{Path, PathBuf},
};

struct Node {
    name: OsString,
    children: HashMap<OsString, Node>,
    is_dir: bool,
}

pub struct Tree {
    root_parent: Option<PathBuf>,
    root: Node,
}

impl Tree {
    pub fn new<T: AsRef<Path>>(root_path: T) -> Result<Self> {
        let root_path = root_path.as_ref();
        assert!(root_path.is_dir());
        let root_parent = root_path.parent().map(|path| PathBuf::from(path));
        let root = explore_path(root_path)?;
        Ok(Self { root_parent, root })
    }
}

fn explore_path(path: &Path) -> Result<Node> {
    let curr_name = OsString::from(path.file_name().unwrap_or_default());
    let mut children = HashMap::new();
    if path.is_dir() {
        for entry in fs::read_dir(path)? {
            let entry = entry?;
            let path = entry.path();
            let node = explore_path(&path)?;
            let curr_name = OsString::from(path.file_name().unwrap_or_default());
            children.insert(curr_name, node);
        }
    }
    Ok(Node {
        name: curr_name,
        children,
        is_dir: path.is_dir(),
    })
}
