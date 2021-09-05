use std::{
    fmt::{self, Display, Formatter},
    path::{Path, PathBuf},
};

#[derive(PartialEq, Eq)]
pub struct Change {
    path: PathBuf,
    kind: ChangeKind,
    is_dir: bool,
}

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum ChangeKind {
    Add,
    Delete,
    Modify,
}

impl Display for ChangeKind {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        use ChangeKind::*;
        match self {
            Add => write!(f, "Add"),
            Delete => write!(f, "Delete"),
            Modify => write!(f, "Modify"),
        }
    }
}

impl Display for Change {
    fn fmt(&self, f: &mut Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {}", self.kind, self.path.display())
    }
}

impl Change {
    pub fn new(path: PathBuf, kind: ChangeKind, is_dir: bool) -> Self {
        Self { path, kind, is_dir }
    }

    pub fn path(&self) -> &Path {
        &self.path
    }

    pub fn kind(&self) -> ChangeKind {
        self.kind
    }

    pub fn is_dir(&self) -> bool {
        self.is_dir
    }
}
