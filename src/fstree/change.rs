use std::{
    fmt::{self, Display, Formatter},
    path::{Path, PathBuf},
};

pub struct Change {
    path: PathBuf,
    kind: ChangeKind,
}

#[derive(Clone, Copy)]
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
    pub fn new(path: PathBuf, kind: ChangeKind) -> Self {
        Self { path, kind }
    }

    pub fn path(&self) -> &Path {
        &self.path
    }

    pub fn kind(&self) -> &ChangeKind {
        &self.kind
    }
}
