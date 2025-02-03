package components

import (
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
)

func RetrieveRepository(baseURL string) (ocm.Repository, error) {
	octx := ocm.DefaultContext()
	repoSpec := ocireg.NewRepositorySpec(baseURL)
	return octx.RepositoryForSpec(repoSpec)
}

func RetrieveComponentVersion(repo ocm.Repository, componentName, componentVersion string) (ocm.ComponentVersionAccess, error) {
	return repo.LookupComponentVersion(componentName, componentVersion)
}

func RetrieveReferencedComponentVersion(repo ocm.Repository, cv ocm.ComponentVersionAccess, name string) (ocm.ComponentVersionAccess, error) {
	ref, err := cv.GetReference(metav1.NewIdentity(name))
	if err != nil {
		return nil, err
	}
	return repo.LookupComponentVersion(ref.GetComponentName(), ref.GetVersion())
}
