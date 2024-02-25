package main

type Activity struct {
	ID  string
	Tag string
}

type Project struct {
	ID         string
	TimeEntrys []TimeEntry
	Activities []Activity
}

type Projects struct {
	Projects []Project
}

func (p *Projects) GetProject(id string) *Project {
	for i := range p.Projects {
		if p.Projects[i].ID == id {
			return &p.Projects[i]
		}
	}
	return nil
}

func (p *Projects) AddProject(project Project) {
	p.Projects = append(p.Projects, project)
}
