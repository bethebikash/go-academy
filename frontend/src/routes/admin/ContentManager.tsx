import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import {
  BookOpen,
  ChevronDown,
  ChevronRight,
  Edit,
  FilePlus,
  FolderPlus,
  Layers,
  Loader2,
  Plus,
  Trash2,
  FileText,
} from 'lucide-react'
import { apiRequest } from '../../lib/api'
import { Course, LearningPath, Module, Lesson } from '../../types'

export default function ContentManager() {
  const [paths, setPaths] = useState<LearningPath[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Expanded nodes state
  const [expandedPaths, setExpandedPaths] = useState<Record<string, boolean>>({})
  const [expandedCourses, setExpandedCourses] = useState<Record<string, boolean>>({})
  const [expandedModules, setExpandedModules] = useState<Record<string, boolean>>({})

  // Cached children
  const [coursesByPath, setCoursesByPath] = useState<Record<string, Course[]>>({})
  const [modulesByCourse, setModulesByCourse] = useState<Record<string, Module[]>>({})
  const [lessonsByModule, setLessonsByModule] = useState<Record<string, Lesson[]>>({})

  // Modals state
  const [newPathModalOpen, setNewPathModalOpen] = useState(false)
  const [newCourseModalOpen, setNewCourseModalOpen] = useState(false)
  const [newModuleModalOpen, setNewModuleModalOpen] = useState(false)

  const [activePathId, setActivePathId] = useState<string>('')
  const [activeCourseId, setActiveCourseId] = useState<string>('')

  // Form states
  const [formTitle, setFormTitle] = useState('')
  const [formSlug, setFormSlug] = useState('')
  const [formDesc, setFormDesc] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const loadPaths = async () => {
    try {
      setLoading(true)
      const data = await apiRequest<LearningPath[]>('/admin/learning-paths')
      setPaths(data || [])
    } catch (err: any) {
      setError(err.message || 'Failed to load paths')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadPaths()
  }, [])

  const togglePath = async (pathId: string) => {
    const isExpanded = expandedPaths[pathId]
    setExpandedPaths((prev) => ({ ...prev, [pathId]: !isExpanded }))

    if (!isExpanded && !coursesByPath[pathId]) {
      try {
        const courses = await apiRequest<Course[]>(`/admin/learning-paths/${pathId}/courses`)
        setCoursesByPath((prev) => ({ ...prev, [pathId]: courses || [] }))
      } catch (err) {
        console.error('Failed to load courses', err)
      }
    }
  }

  const toggleCourse = async (courseId: string) => {
    const isExpanded = expandedCourses[courseId]
    setExpandedCourses((prev) => ({ ...prev, [courseId]: !isExpanded }))

    if (!isExpanded && !modulesByCourse[courseId]) {
      try {
        const modules = await apiRequest<Module[]>(`/admin/courses/${courseId}/modules`)
        setModulesByCourse((prev) => ({ ...prev, [courseId]: modules || [] }))
      } catch (err) {
        console.error('Failed to load modules', err)
      }
    }
  }

  const toggleModule = async (moduleId: string) => {
    const isExpanded = expandedModules[moduleId]
    setExpandedModules((prev) => ({ ...prev, [moduleId]: !isExpanded }))

    if (!isExpanded && !lessonsByModule[moduleId]) {
      try {
        const lessons = await apiRequest<Lesson[]>(`/admin/modules/${moduleId}/lessons`)
        setLessonsByModule((prev) => ({ ...prev, [moduleId]: lessons || [] }))
      } catch (err) {
        console.error('Failed to load lessons', err)
      }
    }
  }

  // Path Creation
  const handleCreatePath = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formTitle) return

    try {
      setSubmitting(true)
      await apiRequest('/admin/learning-paths', {
        method: 'POST',
        body: JSON.stringify({
          title: formTitle,
          slug: formSlug || undefined,
          description: formDesc,
          status: 'draft',
        }),
      })
      setNewPathModalOpen(false)
      setFormTitle('')
      setFormSlug('')
      setFormDesc('')
      await loadPaths()
    } catch (err: any) {
      setError(err.message || 'Failed to create path')
    } finally {
      setSubmitting(false)
    }
  }

  // Course Creation
  const handleCreateCourse = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formTitle || !activePathId) return

    try {
      setSubmitting(true)
      await apiRequest('/admin/courses', {
        method: 'POST',
        body: JSON.stringify({
          learning_path_id: activePathId,
          title: formTitle,
          slug: formSlug || undefined,
          description: formDesc,
          status: 'draft',
        }),
      })
      setNewCourseModalOpen(false)
      setFormTitle('')
      setFormSlug('')
      setFormDesc('')
      // Reload courses for this path
      const courses = await apiRequest<Course[]>(`/admin/learning-paths/${activePathId}/courses`)
      setCoursesByPath((prev) => ({ ...prev, [activePathId]: courses || [] }))
    } catch (err: any) {
      setError(err.message || 'Failed to create course')
    } finally {
      setSubmitting(false)
    }
  }

  // Module Creation
  const handleCreateModule = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formTitle || !activeCourseId) return

    try {
      setSubmitting(true)
      await apiRequest('/admin/modules', {
        method: 'POST',
        body: JSON.stringify({
          course_id: activeCourseId,
          title: formTitle,
          slug: formSlug || undefined,
          description: formDesc,
        }),
      })
      setNewModuleModalOpen(false)
      setFormTitle('')
      setFormSlug('')
      setFormDesc('')
      // Reload modules for this course
      const modules = await apiRequest<Module[]>(`/admin/courses/${activeCourseId}/modules`)
      setModulesByCourse((prev) => ({ ...prev, [activeCourseId]: modules || [] }))
    } catch (err: any) {
      setError(err.message || 'Failed to create module')
    } finally {
      setSubmitting(false)
    }
  }

  // Status toggle
  const handleStatusToggle = async (type: 'path' | 'course' | 'lesson', id: string, currentStatus: string, parentId?: string) => {
    const nextStatus = currentStatus === 'published' ? 'archived' : currentStatus === 'archived' ? 'draft' : 'published'
    try {
      if (type === 'path') {
        await apiRequest(`/admin/learning-paths/${id}/status`, {
          method: 'PATCH',
          body: JSON.stringify({ status: nextStatus }),
        })
        loadPaths()
      } else if (type === 'course') {
        await apiRequest(`/admin/courses/${id}/status`, {
          method: 'PATCH',
          body: JSON.stringify({ status: nextStatus }),
        })
        if (parentId) {
          const courses = await apiRequest<Course[]>(`/admin/learning-paths/${parentId}/courses`)
          setCoursesByPath((prev) => ({ ...prev, [parentId]: courses || [] }))
        }
      } else if (type === 'lesson') {
        await apiRequest(`/admin/lessons/${id}/status`, {
          method: 'PATCH',
          body: JSON.stringify({ status: nextStatus }),
        })
        if (parentId) {
          const lessons = await apiRequest<Lesson[]>(`/admin/modules/${parentId}/lessons`)
          setLessonsByModule((prev) => ({ ...prev, [parentId]: lessons || [] }))
        }
      }
    } catch (err: any) {
      setError(err.message || 'Failed to update status')
    }
  }

  // Delete handlers
  const handleDelete = async (type: 'path' | 'course' | 'module' | 'lesson', id: string, parentId?: string) => {
    if (!confirm(`Are you sure you want to delete this ${type}? This will remove all nested content.`)) return

    try {
      if (type === 'path') {
        await apiRequest(`/admin/learning-paths/${id}`, { method: 'DELETE' })
        loadPaths()
      } else if (type === 'course') {
        await apiRequest(`/admin/courses/${id}`, { method: 'DELETE' })
        if (parentId) {
          const courses = await apiRequest<Course[]>(`/admin/learning-paths/${parentId}/courses`)
          setCoursesByPath((prev) => ({ ...prev, [parentId]: courses || [] }))
        }
      } else if (type === 'module') {
        await apiRequest(`/admin/modules/${id}`, { method: 'DELETE' })
        if (parentId) {
          const modules = await apiRequest<Module[]>(`/admin/courses/${parentId}/modules`)
          setModulesByCourse((prev) => ({ ...prev, [parentId]: modules || [] }))
        }
      } else if (type === 'lesson') {
        await apiRequest(`/admin/lessons/${id}`, { method: 'DELETE' })
        if (parentId) {
          const lessons = await apiRequest<Lesson[]>(`/admin/modules/${parentId}/lessons`)
          setLessonsByModule((prev) => ({ ...prev, [parentId]: lessons || [] }))
        }
      }
    } catch (err: any) {
      setError(err.message || `Failed to delete ${type}`)
    }
  }

  if (loading) {
    return (
      <div className="flex h-72 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-indigo-500" />
      </div>
    )
  }

  return (
    <div className="space-y-8 max-w-6xl mx-auto pb-16">
      {/* Top Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            <Layers className="h-6 w-6 text-amber-400" />
            Curriculum & Content Management
          </h1>
          <p className="text-xs text-neutral-400 mt-1">
            Manage the complete hierarchical curriculum: Learning Paths &gt; Courses &gt; Modules &gt; Lessons.
          </p>
        </div>

        <button
          onClick={() => {
            setFormTitle('')
            setFormSlug('')
            setFormDesc('')
            setNewPathModalOpen(true)
          }}
          className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white shadow transition"
        >
          <Plus className="h-4 w-4" /> New Learning Path
        </button>
      </div>

      {error && (
        <div className="rounded-xl border border-rose-800/80 bg-rose-950/40 p-4 text-xs text-rose-300">
          {error}
        </div>
      )}

      {/* Curriculum Hierarchy Tree */}
      <div className="space-y-4">
        {paths.length === 0 ? (
          <div className="rounded-xl border border-dashed border-neutral-800 p-12 text-center text-neutral-400 text-sm">
            No learning paths found. Click "New Learning Path" to create your first path.
          </div>
        ) : (
          paths.map((path) => {
            const isPathExpanded = expandedPaths[path.id]
            const courses = coursesByPath[path.id] || []

            return (
              <div
                key={path.id}
                className="rounded-xl border border-neutral-800 bg-neutral-900/40 overflow-hidden shadow-sm"
              >
                {/* Learning Path Level */}
                <div className="flex items-center justify-between p-4 bg-neutral-900/80 border-b border-neutral-800/80">
                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => togglePath(path.id)}
                      className="text-neutral-400 hover:text-white transition"
                    >
                      {isPathExpanded ? (
                        <ChevronDown className="h-5 w-5" />
                      ) : (
                        <ChevronRight className="h-5 w-5" />
                      )}
                    </button>
                    <BookOpen className="h-5 w-5 text-indigo-400" />
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="font-bold text-white text-sm">{path.title}</span>
                        <span className="text-xs font-mono text-neutral-500">/{path.slug}</span>
                      </div>
                      {path.description && (
                        <p className="text-xs text-neutral-400 line-clamp-1">{path.description}</p>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <button
                      onClick={() => handleStatusToggle('path', path.id, path.status)}
                      className={`rounded px-2 py-0.5 text-[11px] font-semibold uppercase tracking-wider transition ${
                        path.status === 'published'
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : path.status === 'archived'
                          ? 'bg-neutral-800 text-neutral-400'
                          : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                      }`}
                    >
                      {path.status}
                    </button>

                    <button
                      onClick={() => {
                        setActivePathId(path.id)
                        setFormTitle('')
                        setFormSlug('')
                        setFormDesc('')
                        setNewCourseModalOpen(true)
                      }}
                      className="inline-flex items-center gap-1 rounded bg-neutral-800 hover:bg-neutral-700 px-2.5 py-1 text-xs font-medium text-neutral-300 hover:text-white transition"
                    >
                      <Plus className="h-3 w-3" /> Add Course
                    </button>

                    <button
                      onClick={() => handleDelete('path', path.id)}
                      className="p-1 rounded text-neutral-500 hover:text-rose-400 transition"
                      title="Delete Path"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>

                {/* Courses List */}
                {isPathExpanded && (
                  <div className="p-4 pl-10 space-y-3 bg-neutral-950/40 divide-y divide-neutral-800/40">
                    {courses.length === 0 ? (
                      <div className="py-2 text-xs text-neutral-500 italic">
                        No courses under this path. Click "Add Course" above.
                      </div>
                    ) : (
                      courses.map((course) => {
                        const isCourseExpanded = expandedCourses[course.id]
                        const modules = modulesByCourse[course.id] || []

                        return (
                          <div key={course.id} className="pt-3 first:pt-0 space-y-2">
                            {/* Course Item */}
                            <div className="flex items-center justify-between">
                              <div className="flex items-center gap-2.5">
                                <button
                                  onClick={() => toggleCourse(course.id)}
                                  className="text-neutral-400 hover:text-white transition"
                                >
                                  {isCourseExpanded ? (
                                    <ChevronDown className="h-4 w-4" />
                                  ) : (
                                    <ChevronRight className="h-4 w-4" />
                                  )}
                                </button>
                                <Layers className="h-4 w-4 text-amber-400" />
                                <span className="font-semibold text-neutral-200 text-sm">
                                  {course.title}
                                </span>
                                <span className="text-[11px] font-mono text-neutral-500">
                                  /{course.slug}
                                </span>
                              </div>

                              <div className="flex items-center gap-2.5">
                                <button
                                  onClick={() =>
                                    handleStatusToggle('course', course.id, course.status, path.id)
                                  }
                                  className={`rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider ${
                                    course.status === 'published'
                                      ? 'text-emerald-400 bg-emerald-500/10'
                                      : course.status === 'archived'
                                      ? 'text-neutral-500 bg-neutral-800'
                                      : 'text-amber-400 bg-amber-500/10'
                                  }`}
                                >
                                  {course.status}
                                </button>

                                <button
                                  onClick={() => {
                                    setActiveCourseId(course.id)
                                    setFormTitle('')
                                    setFormSlug('')
                                    setFormDesc('')
                                    setNewModuleModalOpen(true)
                                  }}
                                  className="inline-flex items-center gap-1 rounded bg-neutral-800 hover:bg-neutral-700 px-2 py-0.5 text-[11px] font-medium text-neutral-300"
                                >
                                  <FolderPlus className="h-3 w-3" /> Add Module
                                </button>

                                <button
                                  onClick={() => handleDelete('course', course.id, path.id)}
                                  className="p-1 text-neutral-500 hover:text-rose-400"
                                >
                                  <Trash2 className="h-3.5 w-3.5" />
                                </button>
                              </div>
                            </div>

                            {/* Modules List */}
                            {isCourseExpanded && (
                              <div className="pl-6 space-y-2 border-l border-neutral-800/80 my-2">
                                {modules.length === 0 ? (
                                  <div className="text-xs text-neutral-500 italic py-1">
                                    No modules. Click "Add Module" above.
                                  </div>
                                ) : (
                                  modules.map((mod) => {
                                    const isModuleExpanded = expandedModules[mod.id]
                                    const lessons = lessonsByModule[mod.id] || []

                                    return (
                                      <div key={mod.id} className="space-y-1.5 py-1">
                                        {/* Module Item */}
                                        <div className="flex items-center justify-between">
                                          <div className="flex items-center gap-2">
                                            <button
                                              onClick={() => toggleModule(mod.id)}
                                              className="text-neutral-400 hover:text-white"
                                            >
                                              {isModuleExpanded ? (
                                                <ChevronDown className="h-3.5 w-3.5" />
                                              ) : (
                                                <ChevronRight className="h-3.5 w-3.5" />
                                              )}
                                            </button>
                                            <span className="text-xs font-semibold text-neutral-300">
                                              Module: {mod.title}
                                            </span>
                                          </div>

                                          <div className="flex items-center gap-2">
                                            <Link
                                              to={`/admin/lessons/new?moduleId=${mod.id}&courseSlug=${course.slug}`}
                                              className="inline-flex items-center gap-1 rounded bg-indigo-600/20 hover:bg-indigo-600 px-2 py-0.5 text-[11px] font-medium text-indigo-300 hover:text-white transition"
                                            >
                                              <FilePlus className="h-3 w-3" /> New Lesson
                                            </Link>

                                            <button
                                              onClick={() => handleDelete('module', mod.id, course.id)}
                                              className="p-1 text-neutral-500 hover:text-rose-400"
                                            >
                                              <Trash2 className="h-3 w-3" />
                                            </button>
                                          </div>
                                        </div>

                                        {/* Lessons List in Module */}
                                        {isModuleExpanded && (
                                          <div className="pl-6 space-y-1 border-l border-neutral-800/60 my-1">
                                            {lessons.length === 0 ? (
                                              <div className="text-[11px] text-neutral-500 italic py-1">
                                                No lessons in this module. Click "New Lesson".
                                              </div>
                                            ) : (
                                              lessons.map((lesson) => (
                                                <div
                                                  key={lesson.id}
                                                  className="flex items-center justify-between py-1 px-2 rounded hover:bg-neutral-900/60 text-xs"
                                                >
                                                  <div className="flex items-center gap-2">
                                                    <FileText className="h-3.5 w-3.5 text-neutral-400" />
                                                    <span className="text-neutral-300 font-medium">
                                                      {lesson.title}
                                                    </span>
                                                    <span className="text-[10px] text-neutral-500 font-mono">
                                                      /{lesson.slug}
                                                    </span>
                                                    <span className="text-[10px] rounded bg-neutral-800 px-1 text-neutral-400 uppercase">
                                                      {lesson.difficulty}
                                                    </span>
                                                  </div>

                                                  <div className="flex items-center gap-2">
                                                    <button
                                                      onClick={() =>
                                                        handleStatusToggle(
                                                          'lesson',
                                                          lesson.id,
                                                          lesson.status,
                                                          mod.id,
                                                        )
                                                      }
                                                      className={`rounded px-1 text-[10px] uppercase font-semibold ${
                                                        lesson.status === 'published'
                                                          ? 'text-emerald-400 bg-emerald-500/10'
                                                          : lesson.status === 'archived'
                                                          ? 'text-neutral-500 bg-neutral-800'
                                                          : 'text-amber-400 bg-amber-500/10'
                                                      }`}
                                                    >
                                                      {lesson.status}
                                                    </button>

                                                    <Link
                                                      to={`/admin/lessons/${lesson.id}/edit`}
                                                      className="p-1 text-neutral-400 hover:text-indigo-400"
                                                      title="Edit Lesson"
                                                    >
                                                      <Edit className="h-3 w-3" />
                                                    </Link>

                                                    <button
                                                      onClick={() =>
                                                        handleDelete('lesson', lesson.id, mod.id)
                                                      }
                                                      className="p-1 text-neutral-500 hover:text-rose-400"
                                                      title="Delete Lesson"
                                                    >
                                                      <Trash2 className="h-3 w-3" />
                                                    </button>
                                                  </div>
                                                </div>
                                              ))
                                            )}
                                          </div>
                                        )}
                                      </div>
                                    )
                                  })
                                )}
                              </div>
                            )}
                          </div>
                        )
                      })
                    )}
                  </div>
                )}
              </div>
            )
          })
        )}
      </div>

      {/* New Path Modal */}
      {newPathModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-900 p-6 space-y-4 shadow-xl">
            <h2 className="text-lg font-bold text-white">Create Learning Path</h2>
            <form onSubmit={handleCreatePath} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-neutral-300">Title</label>
                <input
                  type="text"
                  required
                  value={formTitle}
                  onChange={(e) => setFormTitle(e.target.value)}
                  placeholder="e.g. Go Concurrency Deep Dive"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-neutral-300">Slug (Optional)</label>
                <input
                  type="text"
                  value={formSlug}
                  onChange={(e) => setFormSlug(e.target.value)}
                  placeholder="auto-generated from title if blank"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-neutral-300">Description</label>
                <textarea
                  rows={3}
                  value={formDesc}
                  onChange={(e) => setFormDesc(e.target.value)}
                  placeholder="Summary of this path..."
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setNewPathModalOpen(false)}
                  className="rounded-lg px-4 py-2 text-xs font-semibold text-neutral-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white shadow"
                >
                  {submitting ? 'Creating...' : 'Create Path'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* New Course Modal */}
      {newCourseModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-900 p-6 space-y-4 shadow-xl">
            <h2 className="text-lg font-bold text-white">Add Course to Path</h2>
            <form onSubmit={handleCreateCourse} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-neutral-300">Course Title</label>
                <input
                  type="text"
                  required
                  value={formTitle}
                  onChange={(e) => setFormTitle(e.target.value)}
                  placeholder="e.g. Goroutines & Channels"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-neutral-300">Slug (Optional)</label>
                <input
                  type="text"
                  value={formSlug}
                  onChange={(e) => setFormSlug(e.target.value)}
                  placeholder="auto-generated from title if blank"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-neutral-300">Description</label>
                <textarea
                  rows={3}
                  value={formDesc}
                  onChange={(e) => setFormDesc(e.target.value)}
                  placeholder="Summary of this course..."
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setNewCourseModalOpen(false)}
                  className="rounded-lg px-4 py-2 text-xs font-semibold text-neutral-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white shadow"
                >
                  {submitting ? 'Creating...' : 'Add Course'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* New Module Modal */}
      {newModuleModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-2xl border border-neutral-800 bg-neutral-900 p-6 space-y-4 shadow-xl">
            <h2 className="text-lg font-bold text-white">Add Module to Course</h2>
            <form onSubmit={handleCreateModule} className="space-y-4">
              <div>
                <label className="text-xs font-medium text-neutral-300">Module Title</label>
                <input
                  type="text"
                  required
                  value={formTitle}
                  onChange={(e) => setFormTitle(e.target.value)}
                  placeholder="e.g. Channel Synchronization"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-neutral-300">Slug (Optional)</label>
                <input
                  type="text"
                  value={formSlug}
                  onChange={(e) => setFormSlug(e.target.value)}
                  placeholder="auto-generated from title if blank"
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-xs font-medium text-neutral-300">Description</label>
                <textarea
                  rows={3}
                  value={formDesc}
                  onChange={(e) => setFormDesc(e.target.value)}
                  placeholder="Summary of this module..."
                  className="mt-1 w-full rounded-lg border border-neutral-700 bg-neutral-950 px-3 py-2 text-xs text-white placeholder-neutral-500 focus:border-indigo-500 focus:outline-none"
                />
              </div>
              <div className="flex justify-end gap-3 pt-2">
                <button
                  type="button"
                  onClick={() => setNewModuleModalOpen(false)}
                  className="rounded-lg px-4 py-2 text-xs font-semibold text-neutral-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submitting}
                  className="rounded-lg bg-indigo-600 hover:bg-indigo-500 px-4 py-2 text-xs font-semibold text-white shadow"
                >
                  {submitting ? 'Creating...' : 'Add Module'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  )
}
