use anyhow::Result;
use std::{
    collections::HashMap,
    ffi::OsString,
    fs::{self, File},
    os::unix::prelude::MetadataExt,
    path::{Path, PathBuf},
};

// NodeTime contains the lower 32 bits of
// any time variant.
#[derive(PartialEq, Eq)]
struct NodeTime {
    sec: u32,
    nsec: u32,
}

impl NodeTime {
    fn ctime<T: MetadataExt>(data: &T) -> Self {
        Self {
            sec: lower32(data.ctime()),
            nsec: lower32(data.ctime_nsec()),
        }
    }

    fn mtime<T: MetadataExt>(data: &T) -> Self {
        Self {
            sec: lower32(data.mtime()),
            nsec: lower32(data.mtime_nsec()),
        }
    }
}

fn lower32(val: i64) -> u32 {
    const MASK: i64 = 0xFFFFFFFF;
    (val & MASK) as u32
}

// NodeMetadata contains data to determine
// whether a file/folder has changed. A folder is
// said to have changed when one of its child file or
// child folder has changed.
// This is inspired from the `struct stat_data` from cache.h
// of the `git` source code.
#[derive(PartialEq, Eq)]
struct NodeMetadata {
    ctime: NodeTime,
    mtime: NodeTime,
    dev: u64,
    inode: u64,
    size: u64,
}

impl NodeMetadata {
    fn new<T: MetadataExt>(data: &T) -> Self {
        Self {
            ctime: NodeTime::ctime(data),
            mtime: NodeTime::mtime(data),
            dev: data.dev(),
            inode: data.ino(),
            size: data.size(),
        }
    }
}

struct Node {
    name: OsString,
    children: HashMap<OsString, Node>,
    is_dir: bool,
    metadata: NodeMetadata,
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

    #[cfg(test)]
    pub fn files(&self) -> Vec<PathBuf> {
        let mut path = self.root_parent.clone().unwrap_or(PathBuf::from("/"));
        let mut files = self.root.files(&mut path);
        files.sort();
        files
    }
}

impl Node {
    #[cfg(test)]
    fn files(&self, parent_path: &mut PathBuf) -> Vec<PathBuf> {
        let curr_path = parent_path;
        curr_path.push(&self.name);
        let files = if !self.is_dir {
            vec![curr_path.clone()]
        } else {
            let mut files = Vec::new();
            for child in self.children.values() {
                let mut child_files = child.files(curr_path);
                files.append(&mut child_files);
            }
            files
        };
        curr_path.pop();
        files
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
    let file = File::open(path)?;
    Ok(Node {
        name: curr_name,
        children,
        is_dir: path.is_dir(),
        metadata: NodeMetadata::new(&file.metadata()?),
    })
}

#[cfg(test)]
mod tests {
    use std::env;

    use super::*;

    #[test]
    fn test_tree_construction() {
        let mut dir = PathBuf::from(env::var("CARGO_MANIFEST_DIR").unwrap());
        dir.push("data");
        dir.push("tree_dir");

        let tree = Tree::new(&dir).unwrap();
        let files = tree.files();

        let mut expected_files = Vec::new();
        dir.push("a.dat");
        expected_files.push(dir.clone());
        dir.pop();
        dir.push("b.dat");
        expected_files.push(dir.clone());
        dir.pop();
        dir.push("c");
        dir.push("e.dat");
        expected_files.push(dir.clone());
        dir.pop();
        dir.push("f.dat");
        expected_files.push(dir.clone());
        expected_files.sort();

        assert_eq!(files, expected_files);
    }
}
